package provider

import (
	"G4f/g4f"
	"bytes"
	"encoding/json"
	"errors"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Provider interface {
	CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error)
}

type RetryProvider struct {
	ProviderList        []Provider
	Shuffle             bool
	SingleProviderRetry bool
	MaxRetries          int
}

func (r *RetryProvider) CreateAsyncGenerator(messages Messages, recCh chan string, errCh chan error) {

	var err error

	for _, p := range r.ProviderList {

		nerrCh := make(chan error)
		go p.CreateAsyncGenerator(messages, recCh, nerrCh)
		for {
			select {
			case err = <-nerrCh:
				if errors.Is(err, g4f.StreamCompleted) {
					errCh <- err
					return
				}
				errCh <- g4f.ErrStreamRestart
				break
			}
		}
	}
	errCh <- err
}

type BaseProvider struct {
	BaseUrl               string
	Working               bool
	NeedsAuth             bool
	SupportStream         bool
	SupportGpt35          bool
	SupportGpt4           bool
	SupportMessageHistory bool
	Params                string
	ApiKey                string
	ProxyUrl              string
	UseProxy              bool
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
		for k, v := range p.Header {
			req.Header.Add(k, v)
		}
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
