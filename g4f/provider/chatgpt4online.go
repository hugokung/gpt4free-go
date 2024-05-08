package provider

import (
	"G4f/g4f"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Chatgpt4Online struct {
	BaseProvider
	Wpnonce   string
	ContextID string
}

type Messages []map[string]interface{}

func (c *Chatgpt4Online) CreateAsyncGenerator(messages Messages, recv chan string, errCh chan error) {

	cookies, err := g4f.GetArgsFromBrowser(c.BaseUrl+"/chat/", c.ProxyUrl, 5, false)
	log.Printf("cookies: %v, err: %v\n", cookies, err)
	if err != nil {
		errCh <- err
		return
	}

	header := map[string]string{
		"Content-Type":       "application/json",
		"Referer":            c.BaseUrl,
		"Sec-Ch-Ua":          "\"Not-A.Brand\";v=\"99\", \"Chromium\";v=\"124\"",
		"Sec-Ch-Ua-Mobile":   "?0",
		"sec-ch-ua-model":    "\"\"",
		"Sec-Ch-Ua-Platform": "\"macOS\"",
		"Dnt":                "1",
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	}

	req, err := http.NewRequest("GET", c.BaseUrl+"/chat/", nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	for k, v := range cookies {
		ck := &http.Cookie{Name: k, Value: v}
		req.AddCookie(ck)
	}

	cli := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := cli.Do(req)
	if err != nil {
		errCh <- err
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errCh <- err
		return
	}

	re := regexp.MustCompile(`restNonce&quot;:&quot;(.*?)&quot;`)
	matches := re.FindStringSubmatch(string(respBody))
	if len(matches) == 0 {
		errCh <- errors.New("No nonce found")
		return
	}
	c.Wpnonce = matches[1]

	re = regexp.MustCompile(`contextId&quot;:(.*?),`)
	matches = re.FindStringSubmatch(string(respBody))
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
		"chatId":     g4f.GetRandomString(11),
		"contextId":  c.ContextID,
		"messages":   messages,
		"newMessage": messages[len(messages)-1]["content"],
		"stream":     true,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		errCh <- err
		return
	}
	req, err = http.NewRequest("POST",
		c.BaseUrl+"/wp-json/mwai-ui/v1/chats/submit", bytes.NewBuffer(payload))
	if err != nil {
		errCh <- err
		return
	}
	header["X-Wp-Nonce"] = c.Wpnonce
	header["Accept"] = "text/event-stream"
	for k, v := range header {
		req.Header.Set(k, v)
	}

	for k, v := range cookies {
		ck := &http.Cookie{Name: k, Value: v}
		req.AddCookie(ck)
	}

	var client http.Client
	if c.UseProxy {
		proxyURL, err := url.Parse(c.ProxyUrl)
		if err != nil {
			errCh <- errors.New("a.ProxyUrl format error")
			return
		}
		client = http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Proxy:                 http.ProxyURL(proxyURL),
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second * time.Duration(5),
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		client = http.Client{
			Timeout: time.Second * 10,
		}
	}
	resp, err = client.Do(req)
	if err != nil {
		errCh <- err
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		rawLine, rdErr := reader.ReadString('\n')
		if rdErr != nil {
			if errors.Is(rdErr, io.EOF) {
				errCh <- errors.New("incomplete stream")
				return
			}
			errCh <- rdErr
			return
		}

		if rawLine == "" || rawLine[0] == ':' {
			continue
		}

		if strings.Contains(rawLine, ":") {
			data := strings.SplitN(rawLine, ":", 2)
			data[0], data[1] = strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
			switch data[0] {
			case "data":
				if data[1] == "[DONE]" {
					errCh <- io.EOF
				}
				recv <- data[1]
			default:
				errCh <- errors.New("unexpected type: " + data[0])
				return
			}
		}
	}
}
