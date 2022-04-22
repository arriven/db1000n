package utils

import (
	"context"
	"encoding/json"
	"flag"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
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
func CheckCountryOrFail(logger *zap.Logger, cfg *CountryCheckerConfig, proxyURLs string) string {
	isCountryAllowed, country := CheckCountry(logger, strings.Split(cfg.countryBlackListCSV, ","), cfg.strictCountryCheck, proxyURLs, cfg.maxRetries)
	if !isCountryAllowed {
		logger.Fatal("not an allowed country, exiting", zap.String("country", country))
	}

	return country
}

// CheckCountry checks which country the app is running from and whether it is in the blacklist.
func CheckCountry(logger *zap.Logger, countriesToAvoid []string, strictCountryCheck bool, proxyURLs string, maxFetchRetries int) (bool, string) {
	var (
		country, ip string
		err         error
	)

	counter := Counter{Count: maxFetchRetries}
	backoffController := BackoffController{BackoffConfig: DefaultBackoffConfig()}

	for counter.Next() {
		logger.Info("checking IP address,", zap.Int("iter", counter.iter))

		if country, ip, err = fetchLocationInfo(logger, proxyURLs); err != nil {
			logger.Warn("error fetching location info", zap.Error(err))
			Sleep(context.Background(), backoffController.Increment().GetTimeout())
		} else {
			break
		}
	}

	if err != nil {
		logger.Warn("Failed to check the country info", zap.Int("retries", maxFetchRetries))

		if strictCountryCheck {
			return false, ""
		}

		return true, ""
	}

	logger.Info("location info", zap.String("country", country), zap.String("ip", ip))

	for i := range countriesToAvoid {
		if country == strings.TrimSpace(countriesToAvoid[i]) {
			logger.Warn("you might need to enable VPN.")

			return !strictCountryCheck, country
		}
	}

	return true, country
}

func fetchLocationInfo(logger *zap.Logger, proxyURLs string) (country, ip string, err error) {
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
		logger.Info("proxy config detected, using it to check country")

		proxies := strings.Split(proxyURLs, ",")

		proxy := proxies[rand.Intn(len(proxies))] //nolint:gosec // Cryptographically secure random not required

		logger.Info("using proxy", zap.String("proxy", proxy))

		u, err := url.Parse(proxy)
		if err != nil {
			return "", "", err
		}

		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	ipInfo := struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return "", "", err
	}

	return ipInfo.Country, ipInfo.IP, nil
}
