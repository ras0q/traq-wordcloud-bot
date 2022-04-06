package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/Ras96/traq-wordcloud-bot/pkg/cron"
	"github.com/Ras96/traq-wordcloud-bot/pkg/traqapi"
	"github.com/Ras96/traq-wordcloud-bot/pkg/wordcloud"
)

const (
	accessTokenKey    = "TRAQ_ACCESS_TOKEN"
	trendChannelIDKey = "TRAQ_TREND_CHANNEL_ID"
	dictChannelIDKey  = "TRAQ_DICT_CHANNEL_ID"
	imageName         = "wordcloud.png"
)

var (
	accessToken    = os.Getenv(accessTokenKey)
	trendChannelID = os.Getenv(trendChannelIDKey)
	dictChannelID  = os.Getenv(dictChannelIDKey)
	jst            = time.FixedZone("Asia/Tokyo", 9*60*60)
)

func main() {
	if err := traqapi.Setup(accessToken); err != nil {
		log.Fatal(err)
	}

	f := func() {
		if err := postTodayWordcloudToTraq(trendChannelID, dictChannelID); err != nil {
			log.Println("[ERROR]", err)
		}
	}

	if err := cron.Setup(f, jst); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
}

func postTodayWordcloudToTraq(trendChannelID string, dictChannelID string) error {
	msgs, err := traqapi.GetDailyMessages(jst)
	if err != nil {
		return fmt.Errorf("failed to get daily messages: %w", err)
	}

	voc, err := traqapi.GetVocabularyInDirectoryChannel(dictChannelID)
	if err != nil {
		return fmt.Errorf("failed to get vocabulary: %w", err)
	}

	udic, err := wordcloud.MakeUserDict(voc)
	if err != nil {
		return fmt.Errorf("failed to make user dictionary: %w", err)
	}

	wordMap, img, err := wordcloud.GenerateWordcloud(msgs, udic)
	if err != nil {
		return fmt.Errorf("Error generating wordcloud: %w", err)
	}

	file, err := imageToFile(img, imageName)
	if err != nil {
		return fmt.Errorf("Error converting image to file: %w", err)
	}
	defer file.Close()

	fileID, err := traqapi.PostFile(accessToken, trendChannelID, file)
	if err != nil {
		return fmt.Errorf("Error posting file: %w", err)
	}

	if err := traqapi.PostMessage(
		trendChannelID,
		generateMessageContent(wordMap, fileID),
		true,
	); err != nil {
		return fmt.Errorf("Error posting wordcloud: %w", err)
	}

	return nil
}

func imageToFile(img image.Image, path string) (*os.File, error) {
	p, _ := filepath.Abs(path)

	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("Error creating wordcloud file: %w", err)
	}

	if err := png.Encode(f, img); err != nil {
		return nil, fmt.Errorf("Error encoding wordcloud: %w", err)
	}

	_, _ = f.Seek(0, os.SEEK_SET)

	return f, nil
}

func generateMessageContent(wordMap map[string]int, fileID string) string {
	type kv struct {
		key   string
		value int
	}

	var ss []kv
	for k, v := range wordMap {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].value > ss[j].value
	})

	jstToday := time.
		Now().
		In(jst).
		Format("2006/01/02")

	return fmt.Sprintf(
		"Daily Wordcloud (%s) やんね！\n"+
			":first_place: %s: %d回\n"+
			":second_place: %s: %d回\n"+
			":third_place: %s: %d回\n"+
			"https://q.trap.jp/files/%s\n",
		jstToday,
		ss[0].key, ss[0].value,
		ss[1].key, ss[1].value,
		ss[2].key, ss[2].value,
		fileID,
	)
}
