package updater

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func Run() {
	const path = "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.v0.7.json"

	etag := ""
	lastModified := ""
	retries := 0

	for {
		configURL, err := url.ParseRequestURI(path)
		if err != nil {
			log.Fatal("Unable to parse URI")
		}

		const requestTimeout = 20 * time.Second

		client := http.Client{
			Timeout: requestTimeout,
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, configURL.String(), nil)
		req.Header.Add("If-None-Match", etag)
		req.Header.Add("If-Modified-Since", lastModified)

		resp, err := client.Do(req)
		if err != nil {
			log.Println("Unable to perform HTTP request")

			return
		}

		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			log.Printf("Error fetching config, code %d", resp.StatusCode)

			return
		}

		etag = resp.Header.Get("etag")
		lastModified = resp.Header.Get("last-modified")

		res, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Unable to read body")

			return
		}

		if len(res) > 0 {
			file, err := os.Create("config/config.json")
			if err != nil {
				log.Println("Unable to create config/config.json")

				return
			}

			size, err := io.WriteString(file, string(res))
			if err != nil {
				log.Println("Error while writing to config/config.json")

				return
			}

			defer file.Close()

			log.Printf("Saved config/config.json with size %d", size)
		}

		retries++

		time.Sleep(1 * time.Minute)
	}
}
