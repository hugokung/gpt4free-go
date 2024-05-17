package client

import (
	"G4f/g4f"
	"G4f/g4f/models"
	"G4f/g4f/provider"
)

type Client struct {
	RecCh        chan string
	ErrCh        chan error
	ChatMessages provider.Messages
	Model        models.Model
}

func (c Client) Create(messages provider.Messages, modelType string,
	providerType string, stream bool, proxy string, maxTokens int, ApiKey string,
	igoredWorking bool, ignoreStream bool) (*Client, error) {

	model, ok := models.Str2Model[modelType]
	if !ok {
		return nil, g4f.ErrModelType
	}

	cli := &Client{
		RecCh:        make(chan string),
		ErrCh:        make(chan error),
		ChatMessages: messages,
		Model:        model,
	}
	go cli.Model.BestProvider.CreateAsyncGenerator(messages, cli.RecCh, cli.ErrCh)
	return cli, nil
}
