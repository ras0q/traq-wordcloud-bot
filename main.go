package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/Ras96/traq-wordcloud-bot/pkg/cron"
	"github.com/Ras96/traq-wordcloud-bot/pkg/traqapi"
	"github.com/Ras96/traq-wordcloud-bot/pkg/wordcloud"
)

const (
	accessTokenKey = "TRAQ_ACCESS_TOKEN"
	channelIDKey   = "TRAQ_CHANNEL_ID"
	imageName      = "wordcloud.png"
)

var (
	accessToken = os.Getenv(accessTokenKey)
	channelID   = os.Getenv(channelIDKey)
)

func main() {
	if err := traqapi.Setup(accessToken); err != nil {
		log.Fatal(err)
	}

	f := func() {
		if err := postTodayWordcloudToTraq(channelID); err != nil {
			log.Println("[ERROR]", err)
		}
	}

	if err := cron.Setup(f); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
}

func postTodayWordcloudToTraq(channelID string) error {
	msgs, err := traqapi.GetDailyMessages()
	if err != nil {
		return fmt.Errorf("failed to get daily messages: %w", err)
	}

	wordMap, img, err := wordcloud.GenerateWordcloud(msgs)
	if err != nil {
		return fmt.Errorf("Error generating wordcloud: %w", err)
	}

	fileID, err := traqapi.PostImage(accessToken, img, imageName, channelID)
	if err != nil {
		return fmt.Errorf("Error posting image: %w", err)
	}

	if err := traqapi.PostMessage(
		channelID,
		generateMessageContent(wordMap, fileID),
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
		In(time.FixedZone("Asia/Tokyo", 9*60*60)).
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
