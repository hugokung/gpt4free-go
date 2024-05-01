package main

import (
	"G4f/g4f"
	"G4f/g4f/provider"
	"errors"
	"fmt"
	"io"
)

func main() {
	msg := g4f.ChatMessage{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": "hello !"},
	}
	chatRequst := g4f.ChatRequest{
		Messages:    msg,
		Temperature: 1.0,
		TopP:        1,
		N:           1,
		Stream:      true,
		Model:       "gpt-3.5-turbo",
	}
	p := provider.BaseProvider{
		BaseUrl: "https://api.openai.com",
		ApiKey:  "sk-OsMMq65tXdfOIlTUYtocSL7NCsmA7CerN77OkEv29dODg1EA",
	}
	//res, err := p.CreateCompletion(chatRequst)
	//if err != nil {
	//fmt.Printf("CreateCompletion error: %v\n", err)
	//}
	//fmt.Println(res)
	stream, err := p.CreateCompletionStream(chatRequst)
	if err != nil {
		fmt.Printf("CreateCompletionStream error: %v\n", err)
	} else {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("Stream finished")
				return
			}
			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}
			fmt.Println(response.Choices[0].Delta.Content)
		}
	}
	return
}
