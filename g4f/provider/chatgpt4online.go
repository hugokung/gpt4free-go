package provider

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
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

type Messages []map[string]string

func (c *Chatgpt4Online) CreateAsyncGenerator(messages Messages, recv chan string, errCh chan error) {
	targetUrl := "https://chatgpt4online.org/chat/"
	resp, err := http.Get(targetUrl)
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
	c.Wpnonce = matches[0]

	re = regexp.MustCompile(`contextId&quot;:(.*?),`)
	matches = re.FindStringSubmatch(string(respBody))
	if len(matches) == 0 {
		errCh <- errors.New("No contextId found")
		return
	}
	c.ContextID = matches[0]

	header := map[string]string{
		"accept":                      "*/*",
		"accept-encoding":             "gzip, deflate",
		"accept-language":             "en-US",
		"referer":                     "",
		"sec-ch-ua":                   "\"Google Chrome\";v=\"123\", \"Not:A-Brand\";v=\"8\", \"Chromium\";v=\"123\"",
		"sec-ch-ua-arch":              "\"x86\"",
		"sec-ch-ua-bitness":           "\"64\"",
		"sec-ch-ua-full-version":      "\"123.0.6312.122\"",
		"sec-ch-ua-full-version-list": "\"Google Chrome\";v=\"123.0.6312.122\", \"Not:A-Brand\";v=\"8.0.0.0\", \"Chromium\";v=\"123.0.6312.122\"",
		"sec-ch-ua-mobile":            "?0",
		"sec-ch-ua-model":             "\"\"",
		"sec-ch-ua-platform":          "\"Windows\"",
		"sec-ch-ua-platform-version":  "\"15.0.0\"",
		"sec-fetch-dest":              "empty",
		"sec-fetch-mode":              "cors",
		"sec-fetch-site":              "same-origin",
		"user-agent":                  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	}

	data := map[string]interface{}{
		"bot_id":     "default",
		"customId":   "",
		"session":    "N/A",
		"chatId":     "",
		"contextId":  c.ContextID,
		"messages":   messages,
		"newMessage": messages[len(messages)-1]["content"],
		"newImageId": "",
		"stream":     true,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		errCh <- err
		return
	}
	req, err := http.NewRequest("POST", targetUrl+"/wp-json/mwai-ui/v1/chats/submit", bytes.NewBuffer(payload))
	if err != nil {
		errCh <- err
		return
	}
	header["x-wp-nonce"] = c.Wpnonce

	for k, v := range header {
		req.Header.Set(k, v)
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
