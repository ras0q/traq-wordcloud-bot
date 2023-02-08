package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/ras0q/traq-wordcloud-bot/pkg/converter"
	"github.com/ras0q/traq-wordcloud-bot/pkg/cron"
	"github.com/ras0q/traq-wordcloud-bot/pkg/traqapi"
	"github.com/ras0q/traq-wordcloud-bot/pkg/wordcloud"
)

const (
	accessTokenKey    = "TRAQ_ACCESS_TOKEN"
	trendChannelIDKey = "TRAQ_TREND_CHANNEL_ID"
	dictChannelIDKey  = "TRAQ_DICT_CHANNEL_ID"
	hallOfFameKey     = "TRAQ_HALL_OF_FAME_CHANNEL_ID"
	imageName         = "wordcloud.png"
)

var (
	accessToken         = os.Getenv(accessTokenKey)
	trendChannelID      = os.Getenv(trendChannelIDKey)
	dictChannelID       = os.Getenv(dictChannelIDKey)
	hallOfFameChannelID = os.Getenv(hallOfFameKey)
	jst                 = time.FixedZone("Asia/Tokyo", 9*60*60)
	yearlyMsgs          []string
)

func main() {
	if err := traqapi.Setup(accessToken); err != nil {
		log.Fatal(err)
	}

	msgs, err := getYearlyMessages(jst)
	if err != nil {
		log.Fatal(err)
	}

	yearlyMsgs = msgs

	cm := cron.Map{
		// daily wordcloud
		"50 23 * * *": func() {
			msgs, err := getDailyMessages(time.Now().In(jst), jst)
			if err != nil {
				log.Println("[ERROR]", err)
			}

			yearlyMsgs = append(yearlyMsgs, msgs...)
			log.Println("[INFO] yearlyMsgs is initialized", len(yearlyMsgs))

			if err := postWordcloudToTraq(msgs, trendChannelID, dictChannelID); err != nil {
				log.Println("[ERROR]", err)
			}
		},
		// yearly wordcloud
		"50 23 31 12 *": func() {
			if err := postWordcloudToTraq(yearlyMsgs, trendChannelID, dictChannelID); err != nil {
				log.Println("[ERROR]", err)
			}
		},
	}

	if err := cron.Setup(cm, jst); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
}

func getDailyMessages(date time.Time, loc *time.Location) ([]string, error) {
	date = date.In(loc)

	return traqapi.GetMessages(
		time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 99, loc).UTC(),
		time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UTC(),
	)
}

func getYearlyMessages(loc *time.Location) ([]string, error) {
	now := time.Now().In(loc)
	year := now.Year()
	date := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	msgs := make([]string, 0)

	for date.Before(now) {
		m, err := getDailyMessages(date, loc)
		if err != nil {
			return nil, err
		}

		msgs = append(msgs, m...)
		date = date.AddDate(0, 0, 1)

		time.Sleep(5 * time.Second)
	}

	return msgs, nil
}

func postWordcloudToTraq(msgs []string, trendChannelID string, dictChannelID string) error {
	voc, err := traqapi.GetWordList(dictChannelID)
	if err != nil {
		return fmt.Errorf("failed to get vocabulary: %w", err)
	}

	udic, err := wordcloud.MakeUserDict(voc)
	if err != nil {
		return fmt.Errorf("failed to make user dictionary: %w", err)
	}

	hof, err := traqapi.GetWordList(hallOfFameChannelID)
	if err != nil {
		return fmt.Errorf("failed to get hall of fame: %w", err)
	}

	wordCountMap, err := converter.Messages2WordCountMap(msgs, udic, hof)
	if err != nil {
		return fmt.Errorf("failed to convert messages to word count map: %w", err)
	}

	img, err := wordcloud.GenerateWordcloud(wordCountMap)
	if err != nil {
		return fmt.Errorf("Error generating wordcloud: %w", err)
	}

	file, err := converter.Image2File(img, imageName)
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
		generateMessageContent(wordCountMap, fileID),
		true,
	); err != nil {
		return fmt.Errorf("Error posting wordcloud: %w", err)
	}

	return nil
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
