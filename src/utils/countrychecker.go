package utils

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
)

type CountryCheckerConfig struct {
	countryBlackListCSV string
	strictCountryCheck  bool
}

// NewGlobalConfigWithFlags returns a GlobalConfig initialized with command line flags.
func NewCountryCheckerConfigWithFlags() *CountryCheckerConfig {
	var res CountryCheckerConfig

	flag.StringVar(&res.countryBlackListCSV, "country-list", GetEnvStringDefault("COUNTRY_LIST", "Ukraine"), "comma-separated list of countries")
	flag.BoolVar(&res.strictCountryCheck, "strict-country-check", GetEnvBoolDefault("STRICT_COUNTRY_CHECK", false),
		"enable strict country check; will also exit if IP can't be determined")

	return &res
}

// CheckCountryOrFail checks the country of client origin by IP and exits the program if it is in the blacklist.
func CheckCountryOrFail(cfg *CountryCheckerConfig) string {
	isCountryAllowed, country := CheckCountry(strings.Split(cfg.countryBlackListCSV, ","), cfg.strictCountryCheck)
	if !isCountryAllowed {
		log.Fatalf("%q is not an allowed country, exiting", country)
	}

	return country
}

// CheckCountry checks which country the app is running from and whether it is in the blacklist.
func CheckCountry(countriesToAvoid []string, strictCountryCheck bool) (bool, string) {
	const maxFetchRetries = 3

	var (
		country, ip string
		err         error
	)

	counter := Counter{Count: maxFetchRetries}
	backoffController := NewBackoffController(nil)

	for counter.Next() {
		log.Printf("Checking IP address, attempt #%d", counter.iter)

		if country, ip, err = fetchLocationInfo(); err != nil {
			Sleep(context.Background(), backoffController.Increment().GetTimeout())
		}
	}

	if err != nil {
		if strictCountryCheck {
			log.Printf("Failed to check the country info in %d attempts while in strict mode", maxFetchRetries)

			return false, ""
		}

		return true, ""
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
