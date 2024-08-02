package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/ras0q/traq-wordcloud-bot/pkg/config"
	"github.com/ras0q/traq-wordcloud-bot/pkg/converter"
	"github.com/ras0q/traq-wordcloud-bot/pkg/cron"
	"github.com/ras0q/traq-wordcloud-bot/pkg/traqapi"
	"github.com/ras0q/traq-wordcloud-bot/pkg/wordcloud"
)

func main() {
	var (
		runOnce bool
		runDate string
	)
	flag.BoolVar(&runOnce, "once", false, "Run only once, not periodically")
	flag.StringVar(&runDate, "date", time.Now().In(config.JST).Format(time.DateOnly), "Wordcloud date")
	flag.Parse()

	if runOnce {
		date, err := time.Parse(time.DateOnly, runDate)
		if err != nil {
			panic(err)
		}

		msgs, err := getDailyMessages(date)
		if err != nil {
			panic(err)
		}

		if err := postWordcloudToTraq(msgs, date); err != nil {
			panic(err)
		}

		return
	}

	cm := cron.Map{
		// daily wordcloud
		"50 23 * * *": func() {
			today := time.Now().In(config.JST)

			msgs, err := getDailyMessages(today)
			if err != nil {
				log.Println("[ERROR]", err)
			}

			if err := postWordcloudToTraq(msgs, today); err != nil {
				log.Println("[ERROR]", err)
			}
		},
		// yearly wordcloud
		"50 23 31 12 *": func() {
			// TODO: implement
			// if err := postWordcloudToTraq(yearlyMsgs, trendChannelID, dictChannelID); err != nil {
			// 	log.Println("[ERROR]", err)
			// }
		},
	}

	if err := cron.Setup(cm); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
}

func getDailyMessages(date time.Time) ([]string, error) {
	date = date.In(config.JST)

	return traqapi.GetMessages(
		time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 99, config.JST).UTC(),
		time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, config.JST).UTC(),
	)
}

func postWordcloudToTraq(msgs []string, date time.Time) error {
	voc, err := traqapi.GetWordList(config.DictChannelID)
	if err != nil {
		return fmt.Errorf("failed to get vocabulary: %w", err)
	}

	udic, err := wordcloud.MakeUserDict(voc)
	if err != nil {
		return fmt.Errorf("failed to make user dictionary: %w", err)
	}

	hof, err := traqapi.GetWordList(config.HallOfFameChannelID)
	if err != nil {
		return fmt.Errorf("failed to get hall of fame: %w", err)
	}

	wordCountMap, err := converter.Messages2WordCountMap(msgs, udic, hof)
	if err != nil {
		return fmt.Errorf("failed to convert messages to word count map: %w", err)
	}

	// TODO: yearly wordcloudの実装時に戻す
	// if err := db.InsertWordCounts(wordCountMap, date.Format("2006/01/02")); err != nil {
	// 	return fmt.Errorf("failed to insert word counts: %w", err)
	// }

	img, err := wordcloud.GenerateWordcloud(wordCountMap)
	if err != nil {
		return fmt.Errorf("Error generating wordcloud: %w", err)
	}

	file, err := converter.Image2File(img, "wordcloud.png")
	if err != nil {
		return fmt.Errorf("Error converting image to file: %w", err)
	}
	defer file.Close()

	fileID, err := traqapi.PostFile(config.TrendChannelID, file)
	if err != nil {
		return fmt.Errorf("Error posting file: %w", err)
	}

	if err := traqapi.PostMessage(
		config.TrendChannelID,
		generateMessageContent(wordCountMap, fileID, date),
		true,
	); err != nil {
		return fmt.Errorf("Error posting wordcloud: %w", err)
	}

	return nil
}

func generateMessageContent(wordMap map[string]int, fileID string, date time.Time) string {
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

	return fmt.Sprintf(
		"Daily Wordcloud (%s) やんね！\n"+
			":first_place: %s: %d回\n"+
			":second_place: %s: %d回\n"+
			":third_place: %s: %d回\n"+
			"https://q.trap.jp/files/%s\n",
		date.Format("2006/01/02"),
		ss[0].key, ss[0].value,
		ss[1].key, ss[1].value,
		ss[2].key, ss[2].value,
		fileID,
	)
}
