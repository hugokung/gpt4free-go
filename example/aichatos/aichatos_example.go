package main

import (
	"fmt"

	"github.com/hugokung/gpt4free-go/g4f/provider"
)

func main() {
	msg := provider.Messages{
		{"Role": "user", "Content": "你是 gpt-3.5吗?"},
	}

	chat := provider.AiChatOs{
		BaseProvider: &provider.BaseProvider{
			BaseUrl: "https://chat10.aichatos.xyz",
		},
		Api: "https://api.binjie.fun",
	}

	recvCh := make(chan string)
	errCh := make(chan error)

	go chat.CreateAsyncGenerator(msg, recvCh, errCh, "", true, nil)
	for {
		select {
		case res := <-recvCh:
			fmt.Printf("res: %v", res)
			return
		case err := <-errCh:
			fmt.Printf("err: %v", err)
			return
		}
	}
}
