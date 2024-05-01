package provider

import (
	"G4f/g4f"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

type Stream struct {
	reader   *bufio.Reader
	response *http.Response
}

func (s *Stream) Recv() (g4f.ChatStreamResponse, error) {
	for {
		rawLine, rdErr := s.reader.ReadString('\n')
		if rdErr != nil {
			if errors.Is(rdErr, io.EOF) {
				return g4f.ChatStreamResponse{}, errors.New("incomplete stream")
			}
			return g4f.ChatStreamResponse{}, rdErr
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
					return g4f.ChatStreamResponse{}, io.EOF
				}
				var resp g4f.ChatStreamResponse
				err := json.Unmarshal([]byte(data[1]), &resp)
				if err != nil {
					return g4f.ChatStreamResponse{}, err
				}
				return resp, nil
			default:
				return g4f.ChatStreamResponse{}, errors.New("unexpected type: " + data[0])
			}
		}
	}
}

func (a *BaseProvider) CreateCompletionStream(req g4f.ChatRequest) (*Stream, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "text/event-stream",
		"Authorization": "Bearer " + a.ApiKey,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", a.BaseUrl+"/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	var client http.Client
	if a.UseProxy {
		proxyURL, err := url.Parse(a.ProxyUrl)
		if err != nil {
			return nil, errors.New("a.ProxyUrl format error")
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
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("request failed")
	}

	return &Stream{reader: bufio.NewReader(resp.Body), response: resp}, nil
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
