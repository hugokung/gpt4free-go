package provider

import (
	"G4f/g4f"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Provider interface {
	// model, message, stream
	CreateCompletionStream(string, g4f.ChatRequest) (chan string, error)
	CreateCompletion(string, g4f.ChatRequest) (string, error)
}

type RetryProvider interface {
	Provider
}

type BaseProvider struct {
	BaseUrl               string
	Working               bool
	NeedsAuth             bool
	SupportStream         bool
	SupportGpt35          bool
	SupportGpt4           bool
	SupportMessageHistory bool
	Params                string
	ApiKey                string
}

func (a *BaseProvider) CreateCompletionStream(model string, req g4f.ChatRequest) (chan string, error) {
	return nil, nil
}

func (a *BaseProvider) CreateCompletion(model string, req g4f.ChatRequest) (string, error) {

	headers := map[string]string{
		//"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		//"accept":             "*/*",
		//"Connection":      "keep-alive",
		//"accept-language": "en,fr-FR;q=0.9,fr;q=0.8,es-ES;q=0.7,es;q=0.6,en-US;q=0.5,am;q=0.4,de;q=0.3",
		//"cache-control":   "no-cache",
		//"origin":             "https://chatgpt.ai",
		//"pragma":             "no-cache",
		//"sec-ch-ua":          `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`,
		//"sec-ch-ua-mobile":   "?0",
		//"sec-ch-ua-platform": `"Windows"`,
		//"sec-fetch-dest":     "empty",
		//"sec-fetch-mode":     "cors",
		//"sec-fetch-site":     "same-origin",
		//"Content-Length":     "163",
		//"Proxy":         "127.0.0.1:7890",
		//"Verify":        "false",
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + a.ApiKey,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", a.BaseUrl+"/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	proxyURL, err := url.Parse("http://127.0.0.1:7890")
	client := http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy:                 http.ProxyURL(proxyURL),
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * time.Duration(5),
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respData map[string]interface{}
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", err
	}

	data, ok := respData["data"].(string)
	if !ok {
		return "", errors.New("Failed to extract 'data' value from JSON response")
	}

	return data, nil
}
