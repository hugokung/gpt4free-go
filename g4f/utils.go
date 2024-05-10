package g4f

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
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func bypassCloudflare(page playwright.Page) error {

	entries, err := page.Locator("#no-js").All()
	if len(entries) == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	log.Println("cloudflare protected")
	button := page.FrameLocator("iframe").Locator("input")
	if err := button.Click(); err != nil {
		return err
	}
	time.Sleep(3)

	entries, err = page.Locator("#no-js").All()
	if len(entries) != 0 || err != nil {
		if err == nil {
			return errors.New("fail to bypass cloudflare")
		}
		return err
	}
	return nil
}

func GetArgsFromBrowser(tgtUrl string, proxy string, doBypassCloudflare bool) (map[string]string, error) {

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	var browser playwright.Browser
	if proxy != "" {
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Proxy: &playwright.Proxy{
				Server: proxy,
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		browser, err = pw.Chromium.Launch()
		if err != nil {
			return nil, err
		}
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}
	page.OnResponse(func(r playwright.Response) {
		log.Printf("response URL: %v\n response status: %v, response header: %v", r.URL(), r.Status(), r.Headers())
	})
	if _, err := page.Goto(tgtUrl); err != nil {
		return nil, err
	}
	//time.Sleep(3 * time.Second)
	if doBypassCloudflare {
		if doBypassCloudflare {

			log.Println("cloudflare protected")
			button := page.FrameLocator("iframe").Locator("input")
			if err := button.Click(); err != nil {
				return nil, err
			}
			time.Sleep(3 * time.Second)
			tmp, err := button.GetAttribute("type")
			if err != nil {
				return nil, err
			}
			log.Printf("tmp: %s\n", tmp)
			entries, err := page.Locator(".challenge-body-text").All()
			if len(entries) != 0 || err != nil {
				if err == nil {
					return nil, errors.New("fail to bypass cloudflare")
				}
				return nil, err
			}
			time.Sleep(3 * time.Second)
		}
		//cfErr := bypassCloudflare(page)
		//if cfErr != nil {
		//return nil, cfErr
		//}
	}
	ck, err := page.Context().Cookies()
	if err != nil {
		return nil, err
	}

	defer pw.Stop()
	defer browser.Close()

	cookies := map[string]string{}

	for i := range ck {
		cookies[ck[i].Name] = ck[i].Value
	}
	return cookies, nil
}

func GetRandomString(length int) string {
	rd.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rd.Intn(len(chars))]
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

func StreamResponse(resp *http.Response) (recvCh chan string, errCh chan error) {

	reader := bufio.NewReader(resp.Body)
	for {
		rawLine, rdErr := reader.ReadString('\n')
		if rdErr != nil {
			if errors.Is(rdErr, io.EOF) {
				errCh <- errors.New("completed stream")
				return
			}
			errCh <- rdErr
			return
		}

		if rawLine == "" || rawLine[0] == ':' {
			continue
		}

		if strings.Contains(rawLine, ":") {
			data := strings.SplitN(rawLine, ":", 2)
			data[0], data[1] = strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
			switch data[0] {
			case "data":
				if data[1] == "[DONE]" {
					errCh <- io.EOF
					return
				}
				recvCh <- data[1]
			default:
				errCh <- errors.New("unexpected type: " + data[0])
				return
			}
		}
	}
}
