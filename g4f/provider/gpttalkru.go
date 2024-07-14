package provider

import (
	"encoding/json"
	"io"
	"log"

	"github.com/hugokung/gpt4free-go/g4f"
	"github.com/hugokung/gpt4free-go/g4f/utils"
)

type GptTalkRu struct {
	*BaseProvider
}

func (g *GptTalkRu) Create() *GptTalkRu {
	return &GptTalkRu{
		BaseProvider: &BaseProvider{
			BaseUrl:               "https://gpttalk.ru",
			Working:               false,
			NeedsAuth:             false,
			SupportStream:         true,
			SupportGpt35:          true,
			SupportMessageHistory: true,
		},
	}
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

func (g *GptTalkRu) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {

	cookies, err := utils.GetArgsFromBrowser(g.BaseUrl+"/getToken", proxy, true)
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
		Proxy:   proxy,
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
	var respData PublicToken
	err = json.Unmarshal(respBytes, &respData)
	if err != nil {
		errCh <- err
		return
	}

	PublicKey := respData.Response.Key.PublicKey

	log.Printf("PublicKey: %v\n", PublicKey)

	RandomString := utils.GetRandomString(8)
	ShifrText, err := utils.Encrypt(PublicKey, RandomString)
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

	fn := func(content string) (string, error) {
		return content, nil
	}

	utils.StreamResponse(resp, recvCh, errCh, fn)

}
