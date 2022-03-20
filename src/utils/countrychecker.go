package utils

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// CheckCountry allows to check which country the app is running from
func CheckCountry(countriesToAvoid []string, strictCountryCheck bool) (bool, string) {
	const maxFetchRetries = 3

	var country, ip string

	for retries := 1; ; retries++ {
		log.Printf("Checking IP address, attempt #%d", retries)

		var err error
		if country, ip, err = fetchLocationInfo(); err != nil {
			if retries < maxFetchRetries {
				time.Sleep(time.Second)

				continue
			}

			if strictCountryCheck {
				log.Printf("Failed to check the country info in %d attempts while in strict mode", maxFetchRetries)

				return false, ""
			}

			return true, ""
		}

		break
	}

	log.Printf("Current country: %s (%s)", country, ip)

	for i := range countriesToAvoid {
		if country == strings.TrimSpace(countriesToAvoid[i]) {
			log.Println("You might need to enable VPN.")
			openBrowser("https://arriven.github.io/db1000n/vpn/")

			return !strictCountryCheck, country
		}
	}

	return true, country
}

func fetchLocationInfo() (country, ip string, err error) {
	const (
		ipCheckerURI   = "https://api.myip.com/"
		requestTimeout = 3 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipCheckerURI, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")

		return "", "", err
	}

	defer resp.Body.Close()

	ipInfo := struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")

		return "", "", err
	}

	return ipInfo.Country, ipInfo.IP, nil
}
