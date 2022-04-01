package wordcloud

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/psykhi/wordclouds"
)

// wordcloudに含めない単語
var exclusiveWords = []string{
	"人",
	"trap",
	"感じ",
	"あと",
	"ユーザー",
	"自分",
}

func isExclusiveWord(word string) bool {
	for _, w := range exclusiveWords {
		if w == word {
			return true
		}
	}

	return false
}

func GenerateWordcloud(msgs []string) (map[string]int, image.Image, error) {
	wordMap, err := parseToNode(msgs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse to node: %w", err)
	}

	if len(wordMap) == 0 {
		return nil, nil, fmt.Errorf("No wordcloud data")
	}

	wc := wordclouds.NewWordcloud(
		wordMap,
		wordclouds.FontFile("assets/fonts/rounded-l-mplus-2c-medium.ttf"),
		wordclouds.Height(1024),
		wordclouds.Width(1024),
		wordclouds.FontMaxSize(128),
		wordclouds.FontMinSize(8),
		wordclouds.Colors([]color.Color{
			color.RGBA{247, 144, 30, 255},
			color.RGBA{194, 69, 39, 255},
			color.RGBA{38, 103, 118, 255},
			color.RGBA{173, 210, 224, 255},
		}),
	)

	return wordMap, wc.Draw(), nil
}

func parseToNode(msgs []string) (map[string]int, error) {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	wordMap := make(map[string]int)

	for _, msg := range msgs {
		wm := make(map[string]struct{})

		tokens := t.Tokenize(msg)
		for _, token := range tokens {
			fea := token.Features()
			sur := strings.ToLower(token.Surface)

			if fea[0] == "名詞" && fea[1] == "一般" && len(sur) > 1 {
				if _, found := wm[sur]; !found && !isExclusiveWord(sur) {
					wm[sur] = struct{}{}
				}
			}
		}

		for w := range wm {
			wordMap[w]++
		}
	}

	return wordMap, nil
}
