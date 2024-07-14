package models

import "github.com/hugokung/gpt4free-go/g4f/provider"

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
	llama          = provider.Llama{}
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

	Llama2 = Model{
		Name:         "llama2",
		BaseProvider: "",
		BestProvider: &provider.RetryProvider{
			SingleProviderRetry: false,
			MaxRetries:          3,
			IterProvider: &provider.IterProvider{
				Shuffle: false,
				ProviderList: []provider.Provider{
					llama.Create(),
				},
			},
		},
	}
)

func init() {
	Str2Model = map[string]Model{
		"default":       DefaultModel,
		"gpt-3.5-turbo": Gpt35Turbo,
		"llama2":        Llama2,
	}
}
