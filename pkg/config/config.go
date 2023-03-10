package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	AccessToken         = os.Getenv("TRAQ_ACCESS_TOKEN")
	TrendChannelID      = os.Getenv("TRAQ_TREND_CHANNEL_ID")
	DictChannelID       = os.Getenv("TRAQ_DICT_CHANNEL_ID")
	HallOfFameChannelID = os.Getenv("TRAQ_HALL_OF_FAME_CHANNEL_ID")
	JST                 = time.FixedZone("Asia/Tokyo", 9*60*60)
	Mysql               = mysql.Config{
		User:                 os.Getenv("MARIADB_USERNAME"),
		Passwd:               os.Getenv("MARIADB_PASSWORD"),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", os.Getenv("MARIADB_HOSTNAME"), 3306),
		DBName:               os.Getenv("MARIADB_DATABASE"),
		Collation:            "utf8_general_ci",
		Loc:                  JST,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
)
