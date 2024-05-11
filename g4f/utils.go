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
	rd "math/rand"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
)

func GetArgsFromBrowser(tgtUrl string, proxy string, doBypassCloudflare bool) (map[string]string, error) {
	u := launcher.New().Set("disable-blink-features", "AutomationControlled").
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36").Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
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

func StreamResponse(resp *http.Response, recvCh chan string, errCh chan error) {

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
