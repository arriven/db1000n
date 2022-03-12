package utils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}

	log.Printf("Please open %s", url)
}

// CheckCountry allows to check which country the app is running from
func CheckCountry(countriesToAvoid []string) bool {
	type IPInfo struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}

	ipInfo := IPInfo{}

	const ipCheckerURI = "https://api.myip.com/"

	const requestTimeout = 3 * time.Second

	retries := 0
	for ipInfo.IP == "" && retries <= 3 {
		log.Printf("Checking IP address, attempt #%d", retries)

		time.Sleep(1 * time.Second)
		retries++

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipCheckerURI, nil)
		if err != nil {
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")

			continue
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Can't close connection to: %s", ipCheckerURI)

				return
			}
		}()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")

			continue
		}

		err = json.Unmarshal(body, &ipInfo)
		if err != nil {
			log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")

			continue
		}
	}

	if ipInfo.Country == "" {
		return false
	}

	for _, country := range countriesToAvoid {
		if ipInfo.Country == strings.TrimSpace(country) {
			log.Printf("Current country: %s. You might need to enable VPN.", ipInfo.Country)
			openBrowser("https://arriven.github.io/db1000n/vpn/")

			return false
		}
	}

	log.Printf("Current country: %s (%s)", ipInfo.Country, ipInfo.IP)

	return true
}
