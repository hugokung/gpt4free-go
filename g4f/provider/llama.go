package provider

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/hugokung/G4f/g4f"
	util "github.com/hugokung/G4f/g4f/utils"
	"github.com/tidwall/gjson"
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

type tokenAndKey struct {
	Token          string
	IdempotencyKey string
}

func getTokenAndIdempotencyKey(BaseUrl string, proxy string, resCh chan tokenAndKey, errCh chan error) {
	u := launcher.New().Set("disable-blink-features", "AutomationControlled").
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36").Headless(false)
	ut := u.MustLaunch()
	browser := rod.New().ControlURL(ut).NoDefaultDevice().MustConnect()
	page := browser.MustPage("")

	rt := page.HijackRequests()
	var token, tmpKey string
	rt.MustAdd("*", func(h *rod.Hijack) {
		var client *http.Client
		if proxy != "" {
			px, _ := url.Parse(proxy)
			client = &http.Client{
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(px),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		} else {
			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
		}
		//log.Printf("url: %s\n", h.Request.URL())
		//log.Printf("data: %s\n", h.Request.Body())
		if h.Request.URL().String() == BaseUrl+"/api" {
			token = gjson.Get(h.Request.Body(), "token").String()
			tmpKey = gjson.Get(h.Request.Body(), "idempotencyKey").String()
			res := tokenAndKey{
				Token:          token,
				IdempotencyKey: tmpKey,
			}
			resCh <- res
		}
		h.LoadResponse(client, true)
	})
	go rt.Run()

	var iframe *rod.Page
	page.Navigate(BaseUrl)
	utils.Sleep(5)
	tErr := rod.Try(func() {
		iframe = page.Timeout(3 * time.Second).MustElement("iframe").MustFrame().CancelTimeout()
		p := page.Browser().MustPageFromTargetID(proto.TargetTargetID(iframe.FrameID))
		p.MustWaitStable()
		p.Timeout(3 * time.Second).MustElement("input[type=checkbox]").MustClick().CancelTimeout()
		utils.Sleep(3)

		page.MustElement("input[placeholder=\"Send a message\"]").MustInput("hello")
		page.MustElement("button[type=submit]").MustClick()
		utils.Sleep(2)
	})
	if tErr != nil {
		if errors.Is(tErr, context.DeadlineExceeded) {
			return
		} else {
			errCh <- tErr
			return
		}
	}

	defer browser.Close()
	defer page.Close()
	defer rt.Stop()
}

func (l *Llama) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {

	header := g4f.DefaultHeader
	header["Origin"] = l.BaseUrl
	header["Referer"] = l.BaseUrl + "/"
	header["Accept"] = "*/*"
	header["Content-Type"] = "text/plain;charset=UTF-8"
	header["Accept-Language"] = "de,en-US;q=0.7,en;q=0.3"
	header["Accept-Encoding"] = "gzip, deflate, br"

	prompt := LlamaFormatPrompt(messages)
	resCh := make(chan tokenAndKey, 1)
	terrCh := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go getTokenAndIdempotencyKey(l.BaseUrl, proxy, resCh, terrCh)
	var res tokenAndKey
	select {
	case res = <-resCh:
		break
	case err := <-terrCh:
		errCh <- err
		return
	case <-ctx.Done():
		errCh <- errors.New("Get token time out")
		return
	}

	data := map[string]interface{}{
		"prompt":         prompt,
		"model":          l.DefaultModel,
		"systemPrompt":   "You are a helpful assistant.",
		"temperature":    0.75,
		"topP":           0.9,
		"maxTokens":      800,
		"idempotencyKey": res.IdempotencyKey,
		"token":          res.Token,
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

	fn := func(content string) (string, error) {
		return content, nil
	}

	util.StreamResponse(resp, recvCh, errCh, fn)
}
