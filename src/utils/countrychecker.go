package utils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// CheckCountry allows to check which country the app is running from
func CheckCountry(countriesToAvoid []string, strictCountryCheck bool) (bool, string) {
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
		if strictCountryCheck {
			log.Println("Strict country check mode is enabled, exiting")

			return false, ""
		}

		return true, ""
	}

	for _, country := range countriesToAvoid {
		if ipInfo.Country == strings.TrimSpace(country) {
			log.Printf("Current country: %s. You might need to enable VPN.", ipInfo.Country)
			OpenBrowser("https://arriven.github.io/db1000n/vpn/")

			if strictCountryCheck {
				log.Println("Strict country check mode is enabled, exiting")

				return false, ipInfo.Country
			}

			return true, ipInfo.Country
		}
	}

	log.Printf("Current country: %s (%s)", ipInfo.Country, ipInfo.IP)

	return true, ipInfo.Country
}
