package main

import (
	"testing"

	"github.com/Ras96/traq-wordcloud-bot/pkg/traqapi"
	"github.com/Ras96/traq-wordcloud-bot/pkg/wordcloud"
)

func Benchmark_main(b *testing.B) {
	b.ResetTimer()

	if err := traqapi.Setup(accessToken); err != nil {
		b.Fatal(err)
	}

	// TODO: テスト用のメッセージを取得する
	msgs, err := traqapi.GetDailyMessages(jst)
	if err != nil {
		b.Fatal(err)
	}

	_, img, err := wordcloud.GenerateWordcloud(msgs)
	if err != nil {
		b.Fatal(err)
	}

	file, err := imageToFile(img, imageName)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
}
