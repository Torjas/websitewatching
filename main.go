package main

import (
	"bytes"
	"github.com/carlescere/scheduler"
	"github.com/go-rod/rod"
	"go.uber.org/zap"
	"net/http"
	"os"
	"runtime"
)

func main() {

	test()
	logger, _ := zap.NewProduction()

	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	defer sugar.Errorln("Program exited")
	browser := rod.New().MustConnect()

	listings := make(map[string]bool)

	job := func() {
		sugar.Info("New Check")
		defer func() {
			if r := recover(); r != nil {
				sugar.Errorln("Recovered. Error:\n", r)

			}
		}()

		n := false

		page := browser.MustPage(os.Getenv("WATCHER_URL"))

		elist := page.MustElement("#_tb_relevant_results")

		elements := elist.MustElements("li")

		newListings := make(map[string]string)
		foundListings := make(map[string]bool)
		for _, element := range elements {
			header := element.MustElement("h3")
			headerText := header.MustText()

			foundListings[headerText] = true

			_, prs := listings[headerText]
			if !prs {
				n = true
				link := element.MustElement("div > div > div > a.org-but")

				expandurl, err := ExpandUrl("https://inberlinwohnen.de" + *link.MustAttribute("href"))
				if err != nil {
					sugar.Errorw("Error expanding url ", "error", err.Error(), "url", "https://inberlinwohnen.de"+*link.MustAttribute("href"))

				} else {
					newListings[headerText] = expandurl
				}

			}
		}
		sugar.Infoln("Finished checking")
		if n {
			for s, s2 := range newListings {
				sugar.Infow("New Listing", "name", s, "url", s2)

			}
			post(newListings)
		}

		listings = foundListings

	}

	_, err := scheduler.Every(10).Minutes().Run(job)
	if err != nil {
		panic(err)
	}
	// Run now and every X.
	//scheduler.Every(5).Minutes().Run(job)

	// Keep the program from not exiting.
	runtime.Goexit()

}

func post(newlistings map[string]string) {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	text := "Found new listings: \n\n"

	for name, url := range newlistings {
		text += name + "\n<" + url + ">\n\n"
	}

	message := "{\"text\": \"" + text + "\" }"

	var b bytes.Buffer
	b.WriteString(message)
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, os.Getenv("WATCHER_GCHAT_WEBHOOK"), &b)
	if err != nil {
		sugar.Errorw("Error creating request to chat", "error", err.Error(), "message_posted", message)
		println("Error creating request: " + err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("Error posting new listings", "error", err.Error(), "message_posted", message)
	}

	defer resp.Body.Close()
}

func test() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://wikipedia.org", nil)
	if err != nil {
		sugar.Errorw("Error creating test request to wikipedia", "error", err.Error())

	}

	resp, err := client.Do(req)
	if err != nil {
		sugar.Errorw("Error GETting wikipedia content", "error", err.Error())

	}

	defer resp.Body.Close()
}

func ExpandUrl(url string) (string, error) {

	expandedUrl := url

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			expandedUrl = req.URL.String()
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	return expandedUrl, nil
}
