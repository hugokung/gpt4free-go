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
	"strings"
	"time"
)

type ChatMessage []map[string]string

type ChatRequest struct {
	Messages         ChatMessage         `json:"messages"`
	Temperature      float32             `json:"temperature,omitempty"`
	TopP             int                 `json:"top_p,omitempty"`
	N                int                 `json:"n,omitempty"`
	Stream           bool                `json:"stream,omitempty"`
	Max_tokens       int                 `json:"max_tokens,omitempty"`
	PresencePenalty  float32             `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32             `json:"frequency_penalty,omitempty"`
	Model            string              `json:"model"`
	Seed             *int                `json:"seed,omitempty"`
	LogitBias        map[string]int      `json:"logit_bias,omitempty"`
	LogProbs         bool                `json:"logprobs,omitempty"`
	TopLogProbs      int                 `json:"top_logprobs,omitempty"`
	ResponseFormat   *ChatResponseFormat `json:"response_format,omitempty"`
}

type ChatResponseFormatType string

type ChatResponseFormat struct {
	Type ChatResponseFormatType `json:"type,omitempty"`
}

type ChatResponseChoices struct {
	Index        int                  `json:"index"`
	Message      *ChatResponseMessage `json:"message"`
	LogProbs     bool                 `json:"logprobs"`
	FinishReason string               `json:"finish_reason"`
}

type ChatResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type ChatResponse struct {
	ID                string                 `json:"id"`
	Choices           []*ChatResponseChoices `json:"choices"`
	Created           int                    `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Object            string                 `json:"object"`
	Usage             *ChatResponseUsage     `json:"usage"`
}

type ChatStreamResponseFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatStreamResponseToolCalls struct {
	Index    int                         `json:"index"`
	ID       string                      `json:"id"`
	Type     string                      `json:"type"`
	Function *ChatStreamResponseFunction `json:"function"`
}

type ChatStreamResponseDelta struct {
	Content  string                       `json:"content"`
	ToolCall *ChatStreamResponseToolCalls `json:"tool_call"`
	Role     string                       `json:"role"`
}

type ChatStreamTopLogprobs struct {
	Token   string `json:"token"`
	Logprob int    `json:"logprob"`
	Bytes   []int  `json:"bytes"`
}

type ChatStreamLogprobsContent struct {
	Token       string                   `json:"token"`
	Logprob     int                      `json:"logprob"`
	Bytes       []int                    `json:"bytes"`
	TopLogprobs []*ChatStreamTopLogprobs `json:"top_logrobs"`
}

type ChatStreamResponseLogprobs struct {
	Content []*ChatStreamLogprobsContent `json:"content"`
}

type ChatStreamResponseChoices struct {
	Delta        *ChatStreamResponseDelta    `json:"delta"`
	Logprobs     *ChatStreamResponseLogprobs `json:"logprobs"`
	FinishReason string                      `json:"finish_reason"`
	Index        int                         `json:"index"`
}

type ChatStreamResponse struct {
	ID                string                       `json:"id"`
	Choices           []*ChatStreamResponseChoices `json:"choices"`
	Created           int                          `json:"created"`
	Model             string                       `json:"model"`
	SystemFingerprint string                       `json:"system_fingerprint"`
	Object            string                       `json:"object"`
}
type OpenAi struct {
	BaseProvider
}

type Stream struct {
	reader   *bufio.Reader
	response *http.Response
}

func (s *Stream) Recv() (ChatStreamResponse, error) {
	for {
		rawLine, rdErr := s.reader.ReadString('\n')
		if rdErr != nil {
			if errors.Is(rdErr, io.EOF) {
				return ChatStreamResponse{}, errors.New("incomplete stream")
			}
			return ChatStreamResponse{}, rdErr
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
					return ChatStreamResponse{}, io.EOF
				}
				var resp ChatStreamResponse
				err := json.Unmarshal([]byte(data[1]), &resp)
				if err != nil {
					return ChatStreamResponse{}, err
				}
				return resp, nil
			default:
				return ChatStreamResponse{}, errors.New("unexpected type: " + data[0])
			}
		}
	}
}

func (a *OpenAi) CreateCompletionStream(req ChatRequest) (*Stream, error) {
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

func (a *OpenAi) CreateCompletion(req ChatRequest) (ChatResponse, error) {

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + a.ApiKey,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return ChatResponse{}, err
	}

	request, err := http.NewRequest("POST", a.BaseUrl+"/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		return ChatResponse{}, err
	}

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	var client http.Client
	if a.UseProxy {
		proxyURL, err := url.Parse(a.ProxyUrl)
		if err != nil {
			return ChatResponse{}, errors.New("a.ProxyUrl format error")
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
		return ChatResponse{}, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return ChatResponse{}, errors.New("request failed")
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	var respData ChatResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return ChatResponse{}, err
	}

	return respData, nil
}
