package g4f

import (
	"errors"
	"math/rand"
	"time"

	"github.com/playwright-community/playwright-go"
)

func bypassCloudflare(page playwright.Page) error {

	button := page.FrameLocator("iframe").Locator("input")
	if err := button.Click(); err != nil {
		return err
	}
	time.Sleep(3)

	entries, err := page.Locator("#no-js").All()
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

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, err
	}

	ctx, err := browser.NewContext()
	if err != nil {
		return nil, err
	}

	page, err := ctx.NewPage()
	if err != nil {
		return nil, err
	}

	if _, err := page.Goto(tgtUrl); err != nil {
		return nil, err
	}

	if doBypassCloudflare {
		cfErr := bypassCloudflare(page)
		if cfErr != nil {
			return nil, cfErr
		}
	}

	ck, err := ctx.Cookies(tgtUrl)
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
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
