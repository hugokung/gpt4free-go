package g4f

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func GetArgsFromBrowser(tgtUrl string, proxy string, timeout int, doBypassCloudflare bool) (map[string]string, error) {

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
	}
	alloctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(alloctx, chromedp.WithDebugf(log.Printf))
	//cfErr := bypassCloudflare(ctx, tgtUrl)
	//if cfErr != nil {
	//return nil, cfErr
	//}
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
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),       // 将URL替换为目标网站的URL
		chromedp.WaitVisible("body"), // 等待页面加载完成
	); err != nil {
		return err
	}

	// 检查是否检测到Cloudflare保护
	var bodyClass string
	if err := chromedp.Run(ctx, chromedp.Text("body", &bodyClass, chromedp.ByQuery)); err != nil {
		return err
	}
	if bodyClass == "no-js" {
		log.Println("Cloudflare protection detected:", url)

		// 执行JavaScript以打开新标签页
		if err := chromedp.Run(ctx, chromedp.Evaluate(`
            document.getElementById("challenge-body-text").addEventListener('click', function() {
                window.open("`+url+`");
            });
        `, nil)); err != nil {
			log.Println("err1")
			return err
		}
		time.Sleep(5 * time.Second)

		// 在iframe中点击挑战按钮
		var iframes []*cdp.Node

		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#turnstile-wrapper iframe`, chromedp.ByQuery), // 等待iframe加载
			chromedp.Nodes(`#turnstile-wrapper iframe`, &iframes, chromedp.ByQuery),
		); err != nil {
			log.Println("err2")
			return err
		}

		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#challenge-stage input`, chromedp.ByQuery, chromedp.FromNode(iframes[0])), // 等待挑战按钮出现
			chromedp.Click(`#challenge-stage input`, chromedp.ByQuery, chromedp.FromNode(iframes[0])),       // 点击挑战按钮
		); err != nil {
			log.Println("err3")
			return err
		}
	}

	// 等待页面加载完成
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.location.href = window.location.href;`, nil),
		chromedp.WaitVisible("body:not(.no-js)"), // 等待页面加载完成
	); err != nil {
		log.Println("err4")
		return err
	}

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
