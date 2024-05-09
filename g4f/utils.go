package g4f

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/playwright-community/playwright-go"
)

func BypassCfByPlaywright(tgtUrl string) error {

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	log.Printf("run")
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return err
	}
	log.Printf("launch")
	page, err := browser.NewPage()
	if err != nil {
		return err
	}
	log.Printf("page")
	if _, err := page.Goto(tgtUrl); err != nil {
		return err
	}
	log.Printf("11111\n")
	button := page.FrameLocator("iframe").Locator("input")
	//if err := button.WaitFor(); err != nil {
	//return err
	//}
	log.Printf("22222\n")
	t, err1 := button.GetAttribute("type")
	if err1 != nil {
		return err1
	}
	log.Printf("type: %v\n", t)
	if err := button.Click(); err != nil {
		return err
	}
	time.Sleep(3)

	entries, err := page.Locator("#no-js").All()
	if len(entries) != 0 || err != nil {
		log.Printf("can't bypass cloudflare")
		return err
	}
	log.Printf("33333\n")
	return nil
}

func GetArgsFromBrowser(tgtUrl string, proxy string, timeout int, doBypassCloudflare bool) (map[string]string, error) {

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
	}
	alloctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(alloctx, chromedp.WithLogf(log.Printf))
	if doBypassCloudflare {
		cfErr := bypassCloudflare(ctx, tgtUrl)
		//cfErr := BypassCfByPlaywright(tgtUrl)
		if cfErr != nil {
			return nil, cfErr
		}
	}
	//return nil, nil
	defer cancel()
	cookies := map[string]string{}
	err := chromedp.Run(ctx, chromedp.Navigate(tgtUrl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			c, err := network.GetCookies().WithUrls([]string{tgtUrl}).Do(ctx)
			if err != nil {
				return err
			}
			for _, v := range c {
				cookies[v.Name] = v.Value
			}
			return nil
		}))
	_ = chromedp.Cancel(ctx)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}

func bypassCloudflare(ctx context.Context, url string) error {

	// 访问目标页面
	log.Printf("111111\n")
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url), // 将URL替换为目标网站的URL
		//chromedp.WaitVisible(`#challenge-body-text`, chromedp.ByID), // 等待页面加载完成
	); err != nil {
		return err
	}
	log.Printf("22222\n")

	// 检查是否检测到Cloudflare保护
	var bodyClass string
	var ok bool
	if err := chromedp.Run(ctx, chromedp.AttributeValue(`body`, "class", &bodyClass, &ok)); err != nil {
		return err
	}
	if ok && bodyClass == "no-js" {
		log.Println("Cloudflare protection detected:", url)

		// 执行JavaScript以打开新标签页
		if err := chromedp.Run(ctx, chromedp.Evaluate(`
            document.getElementById("challenge-body-text").addEventListener('click', function() {
                window.open("`+url+`");
            });
        `, nil), chromedp.Click(`#challenge-body-text`, chromedp.ByID)); err != nil {
			log.Println("err1")
			return err
		}
		time.Sleep(5 * time.Second)
		log.Printf("33333\n")
		// 在iframe中点击挑战按钮
		//var iframes []*cdp.Node
		script := `
		var iframe = document.querySelector('iframe');
		iframe.contentWindow.document.querySelector('input').click();
		`
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#turnstile-wrapper iframe`, chromedp.ByQuery), // 等待iframe加载
			chromedp.EvaluateAsDevTools(script, nil),
			chromedp.Sleep(1*time.Second),
			//chromedp.Nodes(`#turnstile-wrapper iframe`, &iframes, chromedp.ByQuery),
		); err != nil {
			log.Println("err2")
			return err
		}

		//var btns []*cdp.Node
		//if err := chromedp.Run(ctx,
		//chromedp.Nodes(`#challenge-stage`, &btns, chromedp.ByQuery)); err != nil {
		//log.Println("err31")
		//return err
		//}
		log.Printf("44444\n")
		//log.Printf("4444444, frames: %v, challenge-stage: %v\n", len(iframes), len(btns))
		//var tmp string
		//if err := chromedp.Run(ctx,
		//chromedp.Sleep(5*time.Second),
		//chromedp.AttributeValue(`input`, "type", &tmp, &ok, chromedp.ByQuery, chromedp.FromNode(btns[0])),
		//chromedp.WaitVisible(`#challenge-stage input`, chromedp.ByQuery, chromedp.FromNode(btns[0])), // 等待挑战按钮出现
		//chromedp.Click(`input`, chromedp.ByQuery, chromedp.FromNode(btns[0])),
		//chromedp.Sleep(500*time.Millisecond),
		//); err != nil {
		//log.Println("err3")
		//return err
		//}
		//if ok {
		//log.Printf("%v", tmp)
		//}
	}
	//log.Printf("555555\n")
	// 等待页面加载完成
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.location.href = window.location.href;`, nil),
		chromedp.WaitVisible(`#header-chat`), // 等待页面加载完成
	); err != nil {
		log.Println("err4")
		return err
	}
	log.Printf("6666666\n")
	return nil
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
