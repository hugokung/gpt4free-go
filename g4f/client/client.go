package client

import (
	"errors"
	"time"

	"github.com/hugokung/G4f/g4f"
	"github.com/hugokung/G4f/g4f/models"
	"github.com/hugokung/G4f/g4f/provider"
	"github.com/hugokung/G4f/g4f/utils"
)

type Client struct {
	recCh           chan string
	errCh           chan error
	StreamRespCh    chan provider.ChatStreamResponse
	StreamRespErrCh chan error
	ChatMessages    provider.Messages
}

func (c Client) Create(messages provider.Messages, modelType string,
	providerType string, stream bool, proxy string, maxTokens int, ApiKey string,
	igoredWorking bool, ignoreStream bool) (*Client, error) {

	cli := &Client{
		recCh:           make(chan string),
		errCh:           make(chan error),
		StreamRespCh:    make(chan provider.ChatStreamResponse),
		StreamRespErrCh: make(chan error),
		ChatMessages:    messages,
	}

	params := map[string]interface{}{
		"max_tokens":      maxTokens,
		"api_key":         ApiKey,
		"ignored_working": igoredWorking,
		"ignore_stream":   ignoreStream,
	}

	if providerType != "" {
		pd, ok := provider.Str2Provider[providerType]
		if ok {
			go pd.CreateAsyncGenerator(messages, cli.recCh, cli.errCh, proxy, stream, params)
		}
	}

	if modelType != "" {
		model, ok := models.Str2Model[modelType]
		if !ok {
			model, _ = models.Str2Model["default"]
		}
		go model.BestProvider.CreateAsyncGenerator(messages, cli.recCh, cli.errCh, proxy, stream, params)
	}

	go warpStreamResponse(modelType, cli.recCh, cli.errCh, cli.StreamRespCh, cli.StreamRespErrCh, maxTokens)

	return cli, nil
}

func warpStreamResponse(model string, recCh chan string, errCh chan error,
	streamRespCh chan provider.ChatStreamResponse, streamRespErrCh chan error, maxTokens int) {
	var idx int = 0
	for {
		select {
		case content := <-recCh:
			streamResp := provider.ChatStreamResponse{
				ID:      utils.GetRandomString(28),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Choices: []*provider.ChatStreamResponseChoices{
					&provider.ChatStreamResponseChoices{
						Index:        0,
						FinishReason: "",
						Delta: &provider.ChatStreamResponseDelta{
							Role:    "assistant",
							Content: content,
						},
					},
				},
			}
			if maxTokens != -1 && idx+1 >= maxTokens {
				streamResp.Choices[0].FinishReason = "length"
			}
			idx += 1
			streamRespCh <- streamResp
		case err := <-errCh:
			if errors.Is(err, g4f.ErrStreamRestart) {
				streamRespErrCh <- err
				continue
			}
			streamRespErrCh <- err
			return
		}
	}
}
