package main

import (
	"G4f/g4f"
	"G4f/g4f/provider"
	"fmt"
)

func main() {
	msg := g4f.Message{
		{"role": "user", "content": "hello !"},
	}
	chatRequst := g4f.ChatRequest{
		Messages:    msg,
		Temperature: 1.0,
		Top_p:       1,
		N:           1,
		Stream:      false,
		User:        "user",
	}
	p := provider.BaseProvider{
		BaseUrl: "https://api.openai.com",
		ApiKey:  "sk-OsMMq65tXdfOIlTUYtocSL7NCsmA7CerN77OkEv29dODg1EA",
	}
	res, err := p.CreateCompletion("gpt-3.5-turbo-instruct", chatRequst)
	if err != nil {
		fmt.Printf("CreateCompletion error: %v\n", err)
	}
	fmt.Println(res)
	return
}
