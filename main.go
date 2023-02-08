package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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
	db                  *sqlx.DB
	mysqlConfig         = mysql.Config{
		User:      os.Getenv("MARIADB_USERNAME"),
		Passwd:    os.Getenv("MARIADB_PASSWORD"),
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%d", os.Getenv("MARIADB_HOST"), 3306),
		DBName:    os.Getenv("MARIADB_DATABASE"),
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
		ParseTime: true,
	}
)

type WordCount struct {
	Word  string    `db:"word"`
	Count int       `db:"count"`
	Date  time.Time `db:"date"`
}

func main() {
	if err := traqapi.Setup(accessToken); err != nil {
		log.Fatal(err)
	}

	_db, err := sqlx.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	db = _db

	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS word_count " +
			"(word VARCHAR(255) NOT NULL, count INT NOT NULL, date DATETIME NOT NULL, PRIMARY KEY (word, date))",
	); err != nil {
		log.Fatal(err)
	}

	cm := cron.Map{
		// daily wordcloud
		"50 23 * * *": func() {
			msgs, err := getDailyMessages(time.Now().In(jst), jst)
			if err != nil {
				log.Println("[ERROR]", err)
			}

			if err := postWordcloudToTraq(msgs, trendChannelID, dictChannelID); err != nil {
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

	today := time.Now().In(jst) // TODO: グローバルにする

	wordCounts := make([]*WordCount, 0, len(wordCountMap))
	for word, count := range wordCountMap {
		wordCounts = append(wordCounts, &WordCount{
			Word:  word,
			Count: count,
			Date:  today,
		})
	}

	if _, err := db.NamedExec(
		"INSERT INTO word_counts (word, count, date) "+
			"VALUES (:word, :count, :date) "+
			"ON DUPLICATE KEY UPDATE count = :count",
		wordCounts,
	); err != nil {
		return fmt.Errorf("Error inserting word counts: %w", err)
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
