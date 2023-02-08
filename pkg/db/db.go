package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/ras0q/traq-wordcloud-bot/pkg/config"
)

var Global *sqlx.DB

func init() {
	_db, err := sqlx.Open("mysql", config.Mysql.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	Global = _db

	if _, err := Global.Exec(
		"CREATE TABLE IF NOT EXISTS word_count " +
			"(word VARCHAR(255) NOT NULL, count INT NOT NULL, date DATETIME NOT NULL, PRIMARY KEY (word, date))",
	); err != nil {
		log.Fatal(err)
	}
}
