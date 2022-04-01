package utils

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CountryCheckerConfig struct {
	countryBlackListCSV string
	strictCountryCheck  bool
	maxRetries          int
}

// NewGlobalConfigWithFlags returns a GlobalConfig initialized with command line flags.
func NewCountryCheckerConfigWithFlags() *CountryCheckerConfig {
	const maxFetchRetries = 3

	var res CountryCheckerConfig

	flag.StringVar(&res.countryBlackListCSV, "country-list", GetEnvStringDefault("COUNTRY_LIST", "Ukraine"), "comma-separated list of countries")
	flag.BoolVar(&res.strictCountryCheck, "strict-country-check", GetEnvBoolDefault("STRICT_COUNTRY_CHECK", false),
		"enable strict country check; will also exit if IP can't be determined")
	flag.IntVar(&res.maxRetries, "country-check-retries", GetEnvIntDefault("COUNTRY_CHECK_RETRIES", maxFetchRetries),
		"how much retries should be made when checking the country")

	return &res
}

// CheckCountryOrFail checks the country of client origin by IP and exits the program if it is in the blacklist.
func CheckCountryOrFail(cfg *CountryCheckerConfig, proxyURLs string) string {
	isCountryAllowed, country := CheckCountry(strings.Split(cfg.countryBlackListCSV, ","), cfg.strictCountryCheck, proxyURLs, cfg.maxRetries)
	if !isCountryAllowed {
		log.Fatalf("%q is not an allowed country, exiting", country)
	}

	return country
}

// CheckCountry checks which country the app is running from and whether it is in the blacklist.
func CheckCountry(countriesToAvoid []string, strictCountryCheck bool, proxyURLs string, maxFetchRetries int) (bool, string) {
	var (
		country, ip string
		err         error
	)

	counter := Counter{Count: maxFetchRetries}
	backoffController := BackoffController{BackoffConfig: DefaultBackoffConfig()}

	for counter.Next() {
		log.Printf("Checking IP address, attempt #%d", counter.iter)

		if country, ip, err = fetchLocationInfo(proxyURLs); err != nil {
			Sleep(context.Background(), backoffController.Increment().GetTimeout())
		} else {
			break
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

func fetchLocationInfo(proxyURLs string) (country, ip string, err error) {
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

	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}

	if proxyURLs != "" {
		log.Println("proxy config detected, using it to check country")

		proxies := strings.Split(proxyURLs, ",")

		proxy := proxies[rand.Intn(len(proxies))] //nolint:gosec // Cryptographically secure random not required

		log.Println("using proxy", proxy)

		u, err := url.Parse(proxy)
		if err != nil {
			return "", "", err
		}

		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}}
	}

	resp, err := client.Do(req)
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
