package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/ras0q/traq-wordcloud-bot/pkg/config"
)

var global *sqlx.DB

type wordCount struct {
	Word  string `db:"word"`
	Count int    `db:"count"`
	Date  string `db:"date"`
}

func init() {
	_db, err := sqlx.Open("mysql", config.Mysql.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	global = _db

	if _, err := global.Exec(
		"CREATE TABLE IF NOT EXISTS word_count " +
			"(word VARCHAR(255) NOT NULL, count INT NOT NULL, date CHAR(10) NOT NULL)",
	); err != nil {
		log.Fatal(err)
	}
}

func InsertWordCounts(wordCountMap map[string]int, dateStr string) error {
	wordCounts := make([]*wordCount, 0, len(wordCountMap))
	for word, count := range wordCountMap {
		wordCounts = append(wordCounts, &wordCount{
			Word:  word,
			Count: count,
			Date:  dateStr,
		})
	}

	if _, err := global.NamedExec(
		"INSERT INTO word_counts (word, count, date) "+
			"VALUES (:word, :count, :date) "+
			"ON DUPLICATE KEY UPDATE count = :count",
		wordCounts,
	); err != nil {
		return fmt.Errorf("Error inserting word counts: %w", err)
	}

	return nil
}
