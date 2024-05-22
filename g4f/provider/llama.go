package provider

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/google/uuid"
	"github.com/hugokung/G4f/g4f"
	util "github.com/hugokung/G4f/g4f/utils"
)

type Llama struct {
	*BaseProvider
	DefaultModel string
	Models       []string
	ModelAliases map[string]string
}

func (l *Llama) Create() *Llama {
	return &Llama{
		BaseProvider: &BaseProvider{
			Working:               false,
			SupportMessageHistory: true,
			BaseUrl:               "https://www.llama2.ai",
		},
		DefaultModel: "meta/meta-llama-3-8b-instruct",
		Models: []string{
			"meta/llama-2-7b-chat",
			"meta/llama-2-13b-chat",
			"meta/llama-2-70b-chat",
			"meta/meta-llama-3-8b-instruct",
			"meta/meta-llama-3-70b-instruct",
		},
		ModelAliases: map[string]string{
			"meta-llama/Meta-Llama-3-8B-Instruct":  "meta/meta-llama-3-8b-instruct",
			"meta-llama/Meta-Llama-3-70B-Instruct": "meta/meta-llama-3-70b-instruct",
			"meta-llama/Llama-2-7b-chat-hf":        "meta/llama-2-7b-chat",
			"meta-llama/Llama-2-13b-chat-hf":       "meta/llama-2-13b-chat",
			"meta-llama/Llama-2-70b-chat-hf":       "meta/llama-2-70b-chat",
		},
	}
}

func LlamaFormatPrompt(messages Messages) string {
	prompt := "<|begin_of_text|>"
	if len(messages) <= 1 {
		prompt = "<|begin_of_text|><|start_header_id|>system<|end_header_id|>\nYou are a helpful assistant.<eot_id>\n"
	}
	for _, msg := range messages {
		if msg["role"].(string) != "user" {
			prompt += "<|start_header_id|>system<|end_header_id|>\n"
		} else {
			prompt += "<|start_header_id|>user<|end_header_id|>\n"
		}
		prompt += msg["content"].(string) + "<|eot_id|>\n"
	}
	return prompt
}

func (l *Llama) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {

	u := launcher.New().Set("disable-blink-features", "AutomationControlled").
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36").Headless(false)
	if proxy != "" {
		u = u.Proxy(proxy)
	}
	ut := u.MustLaunch()
	browser := rod.New().ControlURL(ut).MustConnect()

	rt := browser.HijackRequests()
	rt.MustAdd("api", func(h *rod.Hijack) {
		log.Printf("data: %s\n", h.Request.Body())
		h.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go rt.Run()
	page := browser.NoDefaultDevice().MustPage(l.BaseUrl)
	utils.Sleep(5)

	page.MustElement("input").MustInput("hello")
	page.MustElement("#bg-gray-600 hover:bg-gray-800 items-center font-semibold text-white rounded-r-md px-5 py-3").MustClick()
	defer browser.Close()
	defer page.Close()

	header := g4f.DefaultHeader
	header["Origin"] = l.BaseUrl
	header["Referer"] = l.BaseUrl + "/"
	header["Accept"] = "*/*"
	header["Content-Type"] = "text/plain;charset=UTF-8"
	header["Accept-Language"] = "de,en-US;q=0.7,en;q=0.3"
	header["Accept-Encoding"] = "gzip, deflate, br"

	prompt := LlamaFormatPrompt(messages)

	id, err := uuid.NewRandom()
	if err != nil {
		errCh <- err
		return
	}

	data := map[string]interface{}{
		"prompt":         prompt,
		"model":          l.DefaultModel,
		"systemPrompt":   "You are a helpful assistant.",
		"temperature":    0.75,
		"topP":           0.9,
		"maxTokens":      800,
		"idempotencyKey": id.String(),
	}
	client := ProviderHttpClient{
		Header: header,
		Url:    l.BaseUrl + "/api",
		Proxy:  proxy,
		Method: "POST",
		Data:   data,
	}

	resp, err := client.Do()
	if err != nil {
		errCh <- err
		return
	}
	util.StreamResponse(resp, recvCh, errCh, nil)
}
