package models

import "G4f/g4f/provider"

type Model struct {
	Name         string
	BaseProvider string
	BestProvider provider.Provider
}

var (
	gpttalkru      = provider.GptTalkRu{}
	chatgpt4online = provider.Chatgpt4Online{}
	aichatos       = provider.AiChatOs{}
)

var (
	DefaultModel = Model{
		Name:         "",
		BaseProvider: "",
		BestProvider: &provider.RetryProvider{
			Shuffle:             false,
			SingleProviderRetry: true,
			MaxRetries:          3,
			ProviderList: []provider.Provider{
				gpttalkru.Create(),
				chatgpt4online.Create(),
				aichatos.Create(),
			},
		},
	}
)
