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
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/playwright-community/playwright-go"
)

func bypassCloudflare(page playwright.Page) error {
	if title, err := page.Title(); err != nil || title == "Just a moment..." {
		if err != nil {
			return err
		}
		log.Println("cloudflare protected")
		button := page.FrameLocator("iframe").Locator("input[type=checkbox]")
		if err := button.Click(); err != nil {
			return err
		}
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		})
		title, err := page.Title()
		if err != nil || title == "Just a moment..." {
			if err == nil {
				return errors.New("fail to bypass cloudflare")
			}
			return err
		}
	}
	return nil
}

func GetArgsFromRod(tgtUrl string, proxy string) (map[string]string, error) {
	u := launcher.New().Set("disable-blink-features", "AutomationControlled").
		Set("--no-sandbox").
		Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36").Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	page := browser.NoDefaultDevice().MustPage(tgtUrl)
	utils.Sleep(5)
	log.Printf(page.MustInfo().Title)
	iframe := page.MustElement("iframe").MustFrame()
	p := page.Browser().MustPageFromTargetID(proto.TargetTargetID(iframe.FrameID))
	p.MustWaitStable()
	tmp := p.MustElement("input[type=checkbox]").MustAttribute("type")
	log.Printf("tmp: %v\n", tmp)
	p.MustElement("input[type=checkbox]").MustClick()
	//el := iframe.MustElement("input")
	//el.MustClick()

	utils.Sleep(3)
	if page.MustInfo().Title == "Just a moment..." {
		return nil, errors.New("bypass cloudflare failure")
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
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
		})
		if err != nil {
			return nil, err
		}
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}

	script := "Object.defineProperty(navigator, 'webdriver', { get: () => undefined})"
	err = page.AddInitScript(
		playwright.Script{
			Content: &script,
		})
	if err != nil {
		return nil, err
	}

	//page.OnResponse(func(r playwright.Response) {
	//	if r.URL() == tgtUrl {
	//		log.Printf("response status: %v\n, response header: %v\n", r.Status(), r.Headers())
	//	}
	//})
	page.SetDefaultTimeout(60000)
	if _, err := page.Goto(tgtUrl, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return nil, err
	}

	// newPage, err := browser.NewPage()
	// if err != nil {
	// 	return nil, err
	// }
	// err = newPage.AddInitScript(
	// 	playwright.Script{
	// 		Content: &script,
	// 	},
	// )
	// newPage.SetDefaultTimeout(10000)
	// if _, err := newPage.Goto(tgtUrl, playwright.PageGotoOptions{
	// 	WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	// }); err != nil {
	// 	return nil, err
	// }

	//	page.Close()

	if doBypassCloudflare {

		cfErr := bypassCloudflare(page)
		if cfErr != nil {
			return nil, cfErr
		}
	}

	ck, err := page.Context().Cookies()
	if err != nil {
		return nil, err
	}

	defer pw.Stop()
	defer browser.Close()
	defer page.Close()

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
