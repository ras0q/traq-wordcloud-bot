package traqapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/traPtitech/go-traq"
)

var (
	cli  *traq.APIClient
	auth context.Context
)

func Setup(accessToken string) error {
	cli = traq.NewAPIClient(traq.NewConfiguration())
	auth = context.WithValue(context.Background(), traq.ContextAccessToken, accessToken)

	return nil
}

func GetDailyMessages() ([]string, error) {
	var (
		now  = time.Now().UTC()
		r    = regexp.MustCompile(`!\{.+\}|https?:\/\/.+(\s|$)`)
		msgs = make([]string, 0, 5000)
	)

	searchFunc := func(offset int) int {
		_msgs := make([]string, 0, 100)

		res, _, _ := cli.MessageApi.
			SearchMessages(auth).
			Before(time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)).
			After(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)).
			Limit(100).
			Offset(int32(offset * 100)).
			Bot(false).
			Execute()

		for _, msg := range res.Hits {
			if msg.Content != "" {
				plain := r.ReplaceAllString(msg.Content, "")
				_msgs = append(_msgs, plain)
			}
		}

		msgs = append(msgs, _msgs...)

		return int(res.TotalHits)
	}

	// 総メッセージ数を取得するために1回先にAPIを叩く
	totalHits := searchFunc(0)
	num := totalHits / 100 // 並列で回す数

	wg := new(sync.WaitGroup)
	for i := 0; i < num; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			searchFunc(i)
		}(i)
	}
	wg.Wait()

	return msgs, nil
}

func PostImage(accessToken string, img image.Image, path string, channelID string) (string, error) {
	p, _ := filepath.Abs(path)

	file, err := os.Create(p)
	if err != nil {
		return "", fmt.Errorf("Error creating wordcloud file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("Error encoding wordcloud: %w", err)
	}

	_, _ = file.Seek(0, os.SEEK_SET)

	fileID, err := postFile(accessToken, channelID, file)
	if err != nil {
		return "", fmt.Errorf("Error creating wordcloud: %w", err)
	}

	return fileID, nil
}

func postFile(accessToken string, channelID string, file *os.File) (string, error) {
	// NOTE: go-traqがcontent-typeをapplication/octet-streamにしてしまうので自前でAPIを叩く
	// Ref: https://github.com/traPtitech/go-traq/blob/2c7a5f9aa48ef67a6bd6daf4018ca2dabbbbb2f3/client.go#L304
	var b bytes.Buffer
	mw := multipart.NewWriter(&b)

	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Type", "image/png")
	mh.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, file.Name()))

	pw, err := mw.CreatePart(mh)
	if err != nil {
		return "", fmt.Errorf("failed to create part: %w", err)
	}

	if _, err := io.Copy(pw, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	contentType := mw.FormDataContentType()
	mw.Close()

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://q.trap.jp/api/v3/files?channelId=%s", channelID),
		&b,
	)
	if err != nil {
		return "", fmt.Errorf("Error creating request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := new(http.Client)

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)

		return "", fmt.Errorf("Error creating file: %s %s", res.Status, string(b))
	}

	var traqFile traq.FileInfo
	if err := json.NewDecoder(res.Body).Decode(&traqFile); err != nil {
		return "", fmt.Errorf("Error decoding response: %w", err)
	}

	return traqFile.Id, nil
}

func PostMessage(channelID string, content string, embed bool) error {
	_, _, err := cli.MessageApi.
		PostMessage(auth, channelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: content,
			Embed:   &embed,
		}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}
