package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	api2captcha "github.com/2captcha/2captcha-go"

	cdp "github.com/chromedp/chromedp"
)

func wait(sel string) cdp.ActionFunc {
	return run(1*time.Second, cdp.WaitReady(sel))
}

func run(timeout time.Duration, task cdp.Action) cdp.ActionFunc {
	return runFunc(timeout, task.Do)
}

func runFunc(timeout time.Duration, task cdp.ActionFunc) cdp.ActionFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return task.Do(ctx)
	}
}

func solveReCaptcha(client *api2captcha.Client, targetURL, dataSiteKey string) (string, error) {
	c := api2captcha.ReCaptcha{
		SiteKey:   dataSiteKey,
		Url:       targetURL,
		Invisible: true,
		Action:    "verify",
	}

	return client.Solve(c.ToRequest())
}

func recaptchaDemoActions(client *api2captcha.Client) []cdp.Action {
	const targetURL string = "https://www.google.com/recaptcha/api2/demo"
	var siteKey string
	var siteKeyOk bool

	return []cdp.Action{
		run(5*time.Second, cdp.Navigate(targetURL)),
		wait(`[data-sitekey]`),
		wait(`#g-recaptcha-response`),
		wait(`#recaptcha-demo-submit`),
		run(time.Second, cdp.AttributeValue(`[data-sitekey]`, "data-sitekey", &siteKey, &siteKeyOk)),
		runFunc(5*time.Minute, func(ctx context.Context) error {
			if !siteKeyOk {
				return errors.New("missing data-sitekey")
			}

			token, err := solveReCaptcha(client, targetURL, siteKey)
			if err != nil {
				return err
			}

			return cdp.
				SetJavascriptAttribute(`#g-recaptcha-response`, "innerText", token).
				Do(ctx)
		}),
		cdp.Click(`#recaptcha-demo-submit`),
		wait(`.recaptcha-success`),
	}
}

func main() {
	client := api2captcha.NewClient(os.Getenv("API_KEY"))
	actions := recaptchaDemoActions(client)

	opts := append(cdp.DefaultExecAllocatorOptions[:],
		cdp.WindowSize(1366, 768),
		cdp.Flag("headless", false),
		cdp.Flag("incognito", true),
	)

	allocCtx, allocCancel := cdp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()
	ctx, cancel := cdp.NewContext(allocCtx)
	defer cancel()

	start := time.Now()
	err := cdp.Run(ctx, actions...)
	end := time.Since(start)

	if err != nil {
		log.Println("bypass recaptcha failed:", err, end)
	} else {
		log.Println("bypass recaptcha success", end)
	}
}
