package g4f

var DefaultHeader map[string]string

func init() {
	DefaultHeader = map[string]string{
		"Content-Type":       "application/json",
		"Sec-Ch-Ua":          "\"Not-A.Brand\";v=\"99\", \"Chromium\";v=\"124\"",
		"Sec-Ch-Ua-Mobile":   "?0",
		"sec-ch-ua-model":    "\"\"",
		"Sec-Ch-Ua-Platform": "\"macOS\"",
		"Dnt":                "1",
		"User-Agent":         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	}
}
