package provider

import (
	"G4f/g4f"
	"encoding/json"
	"io"
	"log"
)

type GptTalkRu struct {
	BaseProvider
}

type PublicKeyS struct {
	PublicKey string `json:"publicKey"`
}

type TokenResponse struct {
	Key PublicKeyS `json:"key"`
}

type PublicToken struct {
	Response TokenResponse `json:"response"`
}

func (g *GptTalkRu) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error) {

	cookies, err := g4f.GetArgsFromBrowser(g.BaseUrl+"/getToken", g.ProxyUrl, true)
	log.Printf("cookies: %v, err: %v\n", cookies, err)
	if err != nil {
		errCh <- err
		return
	}

	header := g4f.DefaultHeader
	header["accpet"] = "application/json, text/plain, */*"
	client := ProviderHttpClient{
		Header:  header,
		Url:     g.BaseUrl + "/getToken",
		Proxy:   g.ProxyUrl,
		Method:  "GET",
		Cookies: cookies,
	}

	getResp, err := client.Do()

	if err != nil {
		errCh <- err
		return
	}

	defer getResp.Body.Close()
	respBytes, err := io.ReadAll(getResp.Body)
	if err != nil {
		errCh <- err
		return
	}
	//fmt.Printf("getToken Result: %s\n", string(respBytes))
	var respData PublicToken
	err = json.Unmarshal(respBytes, &respData)
	if err != nil {
		errCh <- err
		return
	}

	PublicKey := respData.Response.Key.PublicKey

	log.Printf("PublicKey: %s\n", PublicKey)

	RandomString := g4f.GetRandomString(8)
	ShifrText, err := g4f.Encrypt(PublicKey, RandomString)
	log.Printf("ShifrText: %s\n", ShifrText)
	if err != nil {
		errCh <- err
		return
	}

	data := map[string]interface{}{
		"model":        "",
		"modelType":    1,
		"prompt":       messages,
		"messages":     messages,
		"responseType": "stream",
		"security": map[string]interface{}{
			"randomMessage": RandomString,
			"shifrText":     ShifrText,
		},
	}
	client.Data = data
	client.Method = "POST"
	client.Url = g.BaseUrl + "/gpt4new"
	resp, err := client.Do()
	if err != nil {
		errCh <- err
		return
	}

	recvData, errRes := g4f.StreamResponse(resp)
	for {
		select {
		case res := <-recvData:
			recvCh <- res
		case err = <-errRes:
			errCh <- err
			return
		}
	}
}
