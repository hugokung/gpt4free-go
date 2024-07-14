package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"log"
	rd "math/rand"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/hugokung/gpt4free-go/g4f"
)

func GetArgsFromBrowser(tgtUrl string, proxy string, doBypassCloudflare bool) (map[string]string, error) {
	u := launcher.New().Set("disable-blink-features", "AutomationControlled").
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36").Headless(false)
	if proxy != "" {
		u = u.Proxy(proxy)
	}
	ut := u.MustLaunch()
	browser := rod.New().ControlURL(ut).MustConnect()
	page := browser.NoDefaultDevice().MustPage(tgtUrl)
	utils.Sleep(5)
	if doBypassCloudflare {
		iframe := page.MustElement("iframe").MustFrame()
		p := page.Browser().MustPageFromTargetID(proto.TargetTargetID(iframe.FrameID))
		p.MustWaitStable()
		p.MustElement("input[type=checkbox]").MustClick()
		utils.Sleep(3)
		if page.MustInfo().Title == "Just a moment..." {
			return nil, errors.New("bypass cloudflare failure")
		}
	}
	defer browser.MustClose()
	defer page.MustClose()
	ck, err := page.Cookies([]string{tgtUrl})
	if err != nil {
		return nil, err
	}

	cookies := map[string]string{}
	for i := range ck {
		cookies[ck[i].Name] = ck[i].Value
	}
	return cookies, nil
}

func GetRandomString(length int) string {
	r := rd.New(rd.NewSource(time.Now().UnixNano()))
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func Encrypt(publicKeyPEM, value string) (string, error) {
	// 解码 PEM 格式的公钥
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", errors.New("errInvalidPEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// 将公钥转换为 *rsa.PublicKey 类型
	rsaPubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("errInvalidPublicKey")
	}

	// 使用 RSA 加密数据
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, []byte(value))
	if err != nil {
		return "", err
	}

	// 返回 Base64 编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func StreamResponse(resp *http.Response, recvCh chan string, errCh chan error, fn func(string) (string, error)) {

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Printf("resp status: %v, resp: %v\n", resp.StatusCode, string(respBytes))
		errCh <- g4f.ErrResponse
		return
	}
	reader := bufio.NewReader(resp.Body)
	for {
		rawLine, rdErr := reader.ReadString('\n')
		if rdErr != nil {
			if errors.Is(rdErr, io.EOF) {
				if len(rawLine) != 0 {
					recvCh <- rawLine
					errCh <- g4f.StreamEOF
					return
				}
				errCh <- g4f.StreamEOF
				return
			}
			errCh <- rdErr
			return
		}

		if rawLine == "" || rawLine[0] == ':' {
			continue
		}

		if fn == nil {
			errCh <- errors.New("without decode Funciton")
			return
		}

		decodeData, deErr := fn(rawLine)
		if deErr != nil {
			errCh <- deErr
			if errors.Is(deErr, g4f.StreamCompleted) {
				return
			}
			continue
		}
		recvCh <- decodeData

		//if strings.Contains(rawLine, ":") {
		//	data := strings.SplitN(rawLine, ":", 2)
		//	data[0], data[1] = strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
		//	switch data[0] {
		//	case "data":
		//		if data[1] == "[DONE]" {
		//			errCh <- g4f.StreamCompleted
		//			return
		//		}
		//		log.Printf("data: %v", data[1])
		//		if fn != nil {
		//			decodeData, deErr := fn(data[1])
		//			if deErr != nil {
		//				errCh <- deErr
		//				continue
		//			}
		//			recvCh <- decodeData
		//		} else {
		//			recvCh <- data[1]
		//		}
		//	default:
		//		errCh <- errors.New("unexpected type: " + data[0])
		//		return
		//	}
		//}
	}
}
