package provider

import (
	"errors"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/hugokung/gpt4free-go/g4f"
	"github.com/hugokung/gpt4free-go/g4f/utils"
	"github.com/tidwall/gjson"
)

type Chatgpt4Online struct {
	*BaseProvider
	Wpnonce   string
	ContextID string
}

func (c *Chatgpt4Online) Create() *Chatgpt4Online {
	return &Chatgpt4Online{
		Wpnonce:   "",
		ContextID: "",
		BaseProvider: &BaseProvider{
			BaseUrl:               "https://chatgpt4online.org",
			Working:               true,
			NeedsAuth:             false,
			SupportStream:         true,
			SupportGpt35:          true,
			SupportGpt4:           true,
			SupportMessageHistory: true,
		},
	}
}

func (c *Chatgpt4Online) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {

	cookies, err := utils.GetArgsFromBrowser(c.BaseUrl+"/chat/", proxy, false)
	log.Printf("cookies: %v, err: %v\n", cookies, err)
	if err != nil {
		errCh <- err
		return
	}

	header := g4f.DefaultHeader
	header["Refer"] = c.BaseUrl
	client := ProviderHttpClient{
		Header:  header,
		Url:     c.BaseUrl + "/chat/",
		Proxy:   proxy,
		Method:  "GET",
		Cookies: cookies,
	}
	getResp, err := client.Do()
	if err != nil {
		errCh <- err
		return
	}

	respBytes, err := io.ReadAll(getResp.Body)
	if err != nil {
		errCh <- err
		return
	}
	defer getResp.Body.Close()

	re := regexp.MustCompile(`restNonce&quot;:&quot;(.*?)&quot;`)
	matches := re.FindStringSubmatch(string(respBytes))
	if len(matches) == 0 {
		errCh <- errors.New("No nonce found")
		return
	}
	c.Wpnonce = matches[1]

	re = regexp.MustCompile(`contextId&quot;:(.*?),`)
	matches = re.FindStringSubmatch(string(respBytes))
	if len(matches) == 0 {
		errCh <- errors.New("No contextId found")
		return
	}
	c.ContextID = matches[1]
	log.Printf("nonce: %v, context: %v", c.Wpnonce, c.ContextID)
	data := map[string]interface{}{
		"customId":   "",
		"newFileId":  "",
		"botId":      "default",
		"session":    "N/A",
		"chatId":     utils.GetRandomString(11),
		"contextId":  c.ContextID,
		"messages":   messages,
		"newMessage": messages[len(messages)-1]["content"],
		"stream":     true,
	}

	header["X-Wp-Nonce"] = c.Wpnonce
	header["Accept"] = "text/event-stream"
	client.Data = data
	client.Method = "POST"
	client.Header = header
	client.Url = c.BaseUrl + "/wp-json/mwai-ui/v1/chats/submit"

	resp, err := client.Do()
	if err != nil {
		errCh <- err
		return
	}

	fn := func(content string) (string, error) {
		data := strings.SplitN(content, ":", 2)
		data[0], data[1] = strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
		if data[0] == "data" {
			if data[1] == "[DONE]" {
				return "", g4f.StreamCompleted
			}
			tjson := gjson.Get(content, "type")
			if tjson.String() == "end" {
				return "", g4f.StreamCompleted
			}
			djson := gjson.Get(content, "data")
			return djson.String(), nil
		}
		return "", errors.New(data[0] + ":" + data[1])
	}

	utils.StreamResponse(resp, recvCh, errCh, fn)

}
