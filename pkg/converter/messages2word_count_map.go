package converter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// wordcloudに含めない単語
// TODO: DBで管理する
var exclusiveWords = []string{
	"人",
	"感じ",
	"あと",
	"ー",
}

func isExclusiveWord(word string, hof map[string]struct{}) bool {
	for _, w := range exclusiveWords {
		if strings.EqualFold(w, word) {
			return true
		}
	}

	for w := range hof {
		if strings.EqualFold(w, word) {
			return true
		}
	}

	return false
}

func Messages2WordCountMap(msgs []string, udic *dict.UserDict, hof map[string]struct{}) (map[string]int, error) {
	t, err := tokenizer.New(ipa.DictShrink(), tokenizer.UserDict(udic), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	wordMap := make(map[string]int)
	r := regexp.MustCompile(`!\{.+\}|https?:\/\/.+(\s|$)`)

	for _, msg := range msgs {
		msg := r.ReplaceAllString(msg, "")
		wm := make(map[string]struct{})

		tokens := t.Tokenize(msg)
		for _, token := range tokens {
			fea := token.Features()
			sur := strings.ToLower(token.Surface)

			if (fea[0] == "名詞" && fea[1] == "一般" || fea[0] == "カスタム名詞") && len(sur) > 1 {
				if _, found := wm[sur]; !found && !isExclusiveWord(sur, hof) {
					wm[sur] = struct{}{}
				}
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
