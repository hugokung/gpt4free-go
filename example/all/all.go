package main

import (
	"errors"
	"log"

	"github.com/hugokung/gpt4free-go/g4f"
	"github.com/hugokung/gpt4free-go/g4f/client"
	"github.com/hugokung/gpt4free-go/g4f/provider"
)

func main() {
	msg := provider.Messages{
		{"role": "user", "content": "hello, can you help me?"},
	}
	cli, err := client.Client{}.Create(msg, "gpt-3.5-turbo", "", true, "", 60, "", false, false)
	//cli, err := client.Client{}.Create(msg, "", "gpttalk", true, "", 60, "", false, false)
	if err != nil {
		return
	}

	for {
		select {
		case err := <-cli.StreamRespErrCh:
			log.Printf("cli.ErrCh: %v\n", err)
			if errors.Is(err, g4f.ErrStreamRestart) {
				continue
			}
			return
		case data := <-cli.StreamRespCh:
			log.Printf("recv data: %v\n", data.Choices[0].Delta.Content)
		}
	}
}
