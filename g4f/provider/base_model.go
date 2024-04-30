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
	CreateCompletionStream(g4f.ChatRequest) (g4f.ChatResponse, error)
	CreateCompletion(g4f.ChatRequest) (g4f.ChatResponse, error)
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
	ProxyUrl              string
	UseProxy              bool
}

func (a *BaseProvider) CreateCompletionStream(req g4f.ChatRequest) (chan string, error) {
	return nil, nil
}

func (a *BaseProvider) CreateCompletion(req g4f.ChatRequest) (g4f.ChatResponse, error) {

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + a.ApiKey,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return g4f.ChatResponse{}, err
	}

	request, err := http.NewRequest("POST", a.BaseUrl+"/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		return g4f.ChatResponse{}, err
	}

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	var client http.Client
	if a.UseProxy {
		proxyURL, err := url.Parse(a.ProxyUrl)
		if err != nil {
			return g4f.ChatResponse{}, errors.New("a.ProxyUrl format error")
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
	resp, err := client.Do(request)
	if err != nil {
		return g4f.ChatResponse{}, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return g4f.ChatResponse{}, errors.New("request failed")
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	var respData g4f.ChatResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return g4f.ChatResponse{}, err
	}

	return respData, nil
}
