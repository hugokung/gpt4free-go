package models

import "github.com/hugokung/G4f/g4f/provider"

type Model struct {
	Name         string
	BaseProvider string
	BestProvider provider.Provider
}

var (
	Str2Model      map[string]Model
	gpttalkru      = provider.GptTalkRu{}
	chatgpt4online = provider.Chatgpt4Online{}
	aichatos       = provider.AiChatOs{}
)

var (
	DefaultModel = Model{
		Name:         "",
		BaseProvider: "",
		BestProvider: &provider.RetryProvider{
			SingleProviderRetry: false,
			MaxRetries:          3,
			IterProvider: &provider.IterProvider{
				Shuffle: false,
				ProviderList: []provider.Provider{
					chatgpt4online.Create(),
				},
			},
		},
	}

	Gpt35Turbo = Model{
		Name:         "gpt-3.5-turbo",
		BaseProvider: "openai",
		BestProvider: &provider.RetryProvider{
			SingleProviderRetry: false,
			MaxRetries:          3,
			IterProvider: &provider.IterProvider{
				Shuffle: false,
				ProviderList: []provider.Provider{
					gpttalkru.Create(),
					aichatos.Create(),
				},
			},
		},
	}
)

func init() {
	Str2Model = map[string]Model{
		"default":       DefaultModel,
		"gpt-3.5-turbo": Gpt35Turbo,
	}
}