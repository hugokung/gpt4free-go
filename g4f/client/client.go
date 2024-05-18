package client

import (
	"G4f/g4f"
	"G4f/g4f/models"
	"G4f/g4f/provider"
	"G4f/g4f/utils"
	"time"
)

type Client struct {
	recCh           chan string
	errCh           chan error
	StreamRespCh    chan provider.ChatStreamResponse
	StreamRespErrCh chan error
	ChatMessages    provider.Messages
	Model           models.Model
}

func (c Client) Create(messages provider.Messages, modelType string,
	providerType string, stream bool, proxy string, maxTokens int, ApiKey string,
	igoredWorking bool, ignoreStream bool) (*Client, error) {

	model, ok := models.Str2Model[modelType]
	if !ok {
		return nil, g4f.ErrModelType
	}

	cli := &Client{
		recCh:           make(chan string),
		errCh:           make(chan error),
		StreamRespCh:    make(chan provider.ChatStreamResponse),
		StreamRespErrCh: make(chan error),
		ChatMessages:    messages,
		Model:           model,
	}

	go cli.Model.BestProvider.CreateAsyncGenerator(messages, cli.recCh, cli.errCh)
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
			streamRespErrCh <- err
			return
		}
	}
}
