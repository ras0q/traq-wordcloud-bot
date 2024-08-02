package converter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/filter"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func Messages2WordCountMap(msgs []string, udic *dict.UserDict, hof map[string]struct{}) (map[string]int, error) {
	t, err := tokenizer.New(ipa.DictShrink(), tokenizer.UserDict(udic), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	wordMap := make(map[string]int)
	r := regexp.MustCompile(`!\{.+\}|https?:\/\/.+(\s|$)`)

	var (
		nounFilter           = filter.NewPOSFilter(filter.POS{"名詞", "カスタム名詞"})
		exclusiveWordsFilter = filter.NewWordFilter([]string{"人", "感じ", "あと", "ー"})
	)
	for _, msg := range msgs {
		msg := r.ReplaceAllString(msg, "")
		wm := make(map[string]struct{})

		tokens := t.Tokenize(msg)
		nounFilter.Keep(&tokens)
		exclusiveWordsFilter.Drop(&tokens)

		for _, token := range tokens {
			fea, ok := token.FeatureAt(1)
			if !ok || fea != "一般" {
				continue
			}

			sur := strings.ToLower(token.Surface)
			if _, found := wm[sur]; !found {
				wm[sur] = struct{}{}
			}
		}

		for w := range wm {
			wordMap[w]++
		}
	}

	if len(wordMap) == 0 {
		return nil, fmt.Errorf("No wordcloud data")
	}

	return wordMap, nil
}
