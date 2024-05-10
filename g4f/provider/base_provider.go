package provider

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

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

	var client http.Client
	if p.Proxy != "" {
		proxyURL, err := url.Parse(p.Url)
		if err != nil {
			return nil, err
		}
		client = http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Proxy:                 http.ProxyURL(proxyURL),
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second * time.Duration(5),
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		client = http.Client{
			Timeout: time.Second * 10,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
