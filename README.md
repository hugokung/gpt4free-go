<div align="center">

# GPT4FREE-Go 🆓
This project is only used for learning golang.  
~~Cloudflare is the biggest obstacle~~  
</div>

## ⏬ Installation
```shell
go get github.com/hugokung/G4f
```

## 💡Usage  
#### Text generation  

```golang
msg := provider.Messages{
    {"role": "user", "content": "hello, can you help me?"},
}
cli, err := client.Client{}.Create(msg, "gpt-3.5-turbo", "", true, "", 60, "", false, false)
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
```

## 🚀 TODO
- [ ] Unified API 
- [ ] Docker Deployment

## 🤖 GPT-3.5  
| Website | Provider | Stream | Status | Auth |
| ------  | -------  | ------ | ------ | ---- |
| [chatgpt4online.org](https://chatgpt4online.org) | `g4f.provider.Chatgpt4Online` | ✔️ | ![Unknown](https://img.shields.io/badge/Unknown-grey) | ❌ |
| [gpttalk.ru](https://gpttalk.ru) | `g4f.provider.GptTalkRu` | ✔️ | ![Unknown](https://img.shields.io/badge/Unknown-grey) | ❌ |
| [chat10.aichatos.xyz](https://chat10.aichatos.xyz) | `g4f.provider.AiChatOs` | ✔️ | ![Active](https://img.shields.io/badge/Active-brightgreen) | ❌ |


## ‼️ Declaration  
If you do not want your website to appear here, please raise an issue and I will remove it immediately.
