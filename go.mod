module github.com/ras0q/traq-wordcloud-bot

go 1.22

toolchain go1.22.5

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/ikawaha/kagome-dict v1.1.0
	github.com/ikawaha/kagome-dict/ipa v1.2.0
	github.com/ikawaha/kagome/v2 v2.9.11
	github.com/jmoiron/sqlx v1.4.0
	github.com/psykhi/wordclouds v0.0.0-20231014190151-b9dd58fabbef
	github.com/robfig/cron/v3 v3.0.1
	github.com/traPtitech/go-traq v0.0.0-20240725071454-97c7b85dc879
	golang.org/x/sync v0.7.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
)
