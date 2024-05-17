package provider

import (
	"G4f/g4f"
	"G4f/g4f/utils"
	"errors"
	"io"
	"log"
	"regexp"
)

type Chatgpt4Online struct {
	BaseProvider
	Wpnonce   string
	ContextID string
}

func (c Chatgpt4Online) Create() *Chatgpt4Online {
	return &Chatgpt4Online{
		Wpnonce:   "",
		ContextID: "",
		BaseProvider: BaseProvider{
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

func (c *Chatgpt4Online) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error) {

	cookies, err := utils.GetArgsFromBrowser(c.BaseUrl+"/chat/", c.ProxyUrl, false)
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
		Proxy:   c.ProxyUrl,
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

	utils.StreamResponse(resp, recvCh, errCh)

}
