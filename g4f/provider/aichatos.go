package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hugokung/G4f/g4f"
	"github.com/hugokung/G4f/g4f/utils"
)

type AiChatOs struct {
	*BaseProvider
	Api string
}

func (a *AiChatOs) Create() *AiChatOs {
	return &AiChatOs{
		Api: "https://api.binjie.fun",
		BaseProvider: &BaseProvider{
			BaseUrl:               "https://chat10.aichatos.xyz",
			Working:               true,
			NeedsAuth:             false,
			SupportStream:         true,
			SupportGpt35:          true,
			SupportMessageHistory: true,
		},
	}
}

// FormatPrompt 格式化一系列消息为一个字符串，选项添加特殊标记
func FormatPrompt(messages Messages, addSpecialTokens bool) string {
	if !addSpecialTokens && len(messages) <= 1 {
		if len(messages) == 0 {
			return ""
		}
		msg := messages[0]["content"]
		return msg.(string)
	}

	var formattedMessages []string
	for _, message := range messages {
		formattedMessage := fmt.Sprintf("%s: %s", "User", message["content"].(string))
		formattedMessages = append(formattedMessages, formattedMessage)
	}

	formatted := strings.Join(formattedMessages, "\n")
	return fmt.Sprintf("%s\nAssistant:", formatted)
}

func (a *AiChatOs) CreateAsyncGenerator(messages Messages, recvCh chan string, errCh chan error, proxy string, stream bool, params map[string]interface{}) {
	header := g4f.DefaultHeader
	header["Accept-Language"] = "en-US,en;q=0.5"
	header["Origin"] = a.BaseUrl
	header["Sec-GPC"] = "1"
	header["Connection"] = "keep-alive"
	header["Sec-Fetch-Dest"] = "empty"
	header["Sec-Fetch-Mode"] = "cors"
	header["Sec-Fetch-Site"] = "cross-site"
	header["TE"] = "trailers"

	// 创建一个新的随机数生成器实例并设置种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 生成一个在 [1000000000000, 9999999999999] 范围内的随机整数
	min := int64(1000000000000)
	max := int64(9999999999999)
	userId := r.Int63n(max-min+1) + min
	prompt := FormatPrompt(messages, false)
	data := map[string]interface{}{
		"prompt":         prompt,
		"userId":         fmt.Sprintf("#/chat/%d", userId),
		"network":        true,
		"system":         "",
		"withoutContext": false,
		"stream":         true,
	}
	client := ProviderHttpClient{
		Header: header,
		Url:    a.Api + "/api/generateStream",
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
		data := strings.SplitN(content, ":", 2)
		data[0], data[1] = strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
		if data[0] == "data" {
			if data[1] == "[DONE]" {
				return "", g4f.StreamCompleted
			} else {
				return data[1], nil
			}
		}
		return "", errors.New(data[0] + ":" + data[1])
	}

	utils.StreamResponse(resp, recvCh, errCh, fn)
}
