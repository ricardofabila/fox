package utils

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"

	"github.com/ricardofabila/fox/src/constants"
)

func GetFromAPI(url string) ([]byte, error) {
	httpClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	if res.Body != nil {
		err = res.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

func DownloadFile(url, filename string) error {
	started := time.Now().UnixMilli()
	spin := spinner.New(constants.Clocks, 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	_ = spin.Color("bold", "fgHiYellow")
	spin.Start()

	// So that you get your cursor back
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		spin.Stop()
		color.Yellow(" 😜 Operation aborted")
		os.Exit(1)
	}()

	path := "./" + filename
	// Delete the file manually
	err := RemoveFile(path)
	if err != nil {
		return err
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	// Put content on file
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	spin.Stop()
	ended := time.Now().UnixMilli()
	timeItTook := float64(ended-started) / 1000
	color.Green(" Downloaded in %.2f seconds!", timeItTook)

	return nil
}
