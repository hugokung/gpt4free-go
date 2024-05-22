package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/hugokung/G4f/g4f"
)

var (
	Str2Provider map[string]Provider
)

func init() {
	var (
		aichatos            = &AiChatOs{}
		aichatosRetry       = &RetryProvider{}
		chatgpt4online      = &Chatgpt4Online{}
		chatgpt4onlineRetry = &RetryProvider{}
		gpttalkru           = &GptTalkRu{}
		gpttalkruRetry      = &RetryProvider{}
	)

	aichatosRetry = aichatosRetry.Create()
	chatgpt4onlineRetry = chatgpt4onlineRetry.Create()
	gpttalkruRetry = gpttalkruRetry.Create()

	aichatosRetry.ProviderList = append(aichatosRetry.ProviderList, aichatos.Create())
	chatgpt4onlineRetry.ProviderList = append(chatgpt4onlineRetry.ProviderList, chatgpt4online.Create())
	gpttalkruRetry.ProviderList = append(gpttalkruRetry.ProviderList, gpttalkru.Create())

	Str2Provider = map[string]Provider{
		"aichatos":       aichatosRetry,
		"chatgpt4online": chatgpt4onlineRetry,
		"gpttalk":        gpttalkruRetry,
	}
}

type Provider interface {
	CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{})
}

type IterProvider struct {
	ProviderList []Provider
	Shuffle      bool
}

func (iter *IterProvider) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {
	var err error
	for i, p := range iter.ProviderList {
		nerrCh := make(chan error)
		go p.CreateAsyncGenerator(messages, recvCh, nerrCh, proxy, stream, params)
		for {
			flag := false
			select {
			case err = <-nerrCh:
				if errors.Is(err, g4f.StreamEOF) || errors.Is(err, g4f.StreamCompleted) {
					errCh <- g4f.StreamEOF
					return
				}
				if i == len(iter.ProviderList)-1 {
					errCh <- err
					return
				}
				errCh <- g4f.ErrStreamRestart
				flag = true
				break
			}
			if flag {
				break
			}
		}
		log.Printf("change provider\n")
	}
}

type RetryProvider struct {
	*IterProvider
	SingleProviderRetry bool
	MaxRetries          int
}

func (r *RetryProvider) Create() *RetryProvider {
	return &RetryProvider{
		SingleProviderRetry: true,
		MaxRetries:          3,
		IterProvider: &IterProvider{
			ProviderList: []Provider{},
			Shuffle:      false,
		},
	}
}

func (r *RetryProvider) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {

	var err error
	if r.SingleProviderRetry {
		for i := 0; i < r.MaxRetries; i++ {
			nerrCh := make(chan error)
			go r.ProviderList[0].CreateAsyncGenerator(messages, recvCh, nerrCh, proxy, stream, params)
			for {
				flag := false
				select {
				case err = <-nerrCh:
					if errors.Is(err, g4f.StreamEOF) || errors.Is(err, g4f.StreamCompleted) {
						errCh <- g4f.StreamEOF
						return
					}
					if i == r.MaxRetries-1 {
						errCh <- err
						return
					}
					errCh <- g4f.ErrStreamRestart
					flag = true
					break
				}
				if flag {
					break
				}
			}
			log.Printf("try %d time\n", i+1)
		}
	} else {
		go r.IterProvider.CreateAsyncGenerator(messages, recvCh, errCh, proxy, stream, params)
	}
}

type BaseProvider struct {
	BaseUrl               string
	Working               bool
	NeedsAuth             bool
	SupportStream         bool
	SupportGpt35          bool
	SupportGpt4           bool
	SupportMessageHistory bool
}

type Messages []map[string]interface{}

type ProviderHttpClient struct {
	Header  map[string]string
	Cookies map[string]string
	Data    map[string]interface{}
	Proxy   string
	Method  string
	Url     string
}

func (p *ProviderHttpClient) Do() (*http.Response, error) {

	jar := tls_client.NewCookieJar()
	var options []tls_client.HttpClientOption
	if p.Proxy != "" {
		options = []tls_client.HttpClientOption{
			tls_client.WithTimeoutSeconds(600),
			tls_client.WithClientProfile(profiles.Chrome_124),
			tls_client.WithCookieJar(jar),
			tls_client.WithProxyUrl(p.Proxy),
		}
	} else {
		options = []tls_client.HttpClientOption{
			tls_client.WithTimeoutSeconds(600),
			tls_client.WithClientProfile(profiles.Okhttp4Android10),
			tls_client.WithCookieJar(jar),
		}
	}

	var req *http.Request
	var err error

	if p.Method == "GET" {
		req, err = http.NewRequest("GET", p.Url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		payload, err := json.Marshal(p.Data)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest("POST", p.Url, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
	}

	if p.Header != nil {
		hd := http.Header{
			http.HeaderOrderKey: {
				"Accept",
				"Accept-Encoding",
				"Accept-Language",
				"Content-Type",
				"Dnt",
				"Origin",
				"Referer",
				"Sec-Ch-Ua",
				"Sec-Ch-Ua-Mobile",
				"Sec-Ch-Ua-Platform",
				"Sec-Fetch-Dest",
				"User-Agent",
			},
		}
		for k, v := range p.Header {
			hd[k] = []string{v}
		}
		req.Header = hd
	}

	if p.Cookies != nil {
		for k, v := range p.Cookies {
			cookie := http.Cookie{Name: k, Value: v}
			req.AddCookie(&cookie)
		}
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
