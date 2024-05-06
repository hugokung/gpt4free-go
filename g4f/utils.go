package g4f

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func setheaders(host string, headers map[string]interface{}) chromedp.Tasks {

	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(headers)),
		chromedp.Navigate(host),
	}
}
func GetArgsFromBrowser(tgtUrl string, proxy string, timeout int, doBypassCloudflare bool) (map[string]string, error) {

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
	}
	alloctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(alloctx, chromedp.WithDebugf(log.Printf))
	defer cancel()
	cookies := map[string]string{}
	err := chromedp.Run(ctx, chromedp.Navigate(tgtUrl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			c, err := network.GetCookies().WithUrls([]string{tgtUrl}).Do(ctx)
			if err != nil {
				return err
			}
			for _, v := range c {
				cookies[v.Name] = v.Domain
			}
			return nil
		}))
	_ = chromedp.Cancel(ctx)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}
