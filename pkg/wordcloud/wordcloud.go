package wordcloud

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/psykhi/wordclouds"
)

func MakeUserDict(voc map[string]struct{}) (*dict.UserDict, error) {
	r := make(dict.UserDictRecords, 0, len(voc))

	for k := range voc {
		replaced := strings.TrimSpace(k)
		lines := strings.Split(replaced, "\n")

		if len(lines[0]) == 0 {
			continue
		}

		d := dict.UserDicRecord{
			Text:   lines[0],
			Tokens: []string{lines[0]},
			Yomi:   []string{"ふめい"},
			Pos:    "カスタム名詞",
		}

		if len(lines) > 1 {
			d.Yomi = []string{lines[1]}
		}

		r = append(r, d)
	}

	udic, err := r.NewUserDict()
	if err != nil {
		return nil, fmt.Errorf("failed to create user dict: %w", err)
	}

	return udic, nil
}

func GenerateWordcloud(wordCountMap map[string]int) (image.Image, error) {
	wc := wordclouds.NewWordcloud(
		wordCountMap,
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

	return wc.Draw(), nil
}
