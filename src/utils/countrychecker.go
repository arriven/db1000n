package utils

import (
	"context"
	"encoding/json"
	"flag"
	"net"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
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
func CheckCountryOrFail(logger *zap.Logger, cfg *CountryCheckerConfig, proxyParams ProxyParams) string {
	isCountryAllowed, country := CheckCountry(logger, strings.Split(cfg.countryBlackListCSV, ","), cfg.strictCountryCheck, proxyParams, cfg.maxRetries)
	if !isCountryAllowed {
		logger.Fatal("not an allowed country, exiting", zap.String("country", country))
	}

	return country
}

// CheckCountry checks which country the app is running from and whether it is in the blacklist.
func CheckCountry(logger *zap.Logger, countriesToAvoid []string, strictCountryCheck bool, proxyParams ProxyParams, maxFetchRetries int) (bool, string) {
	var (
		country, ip string
		err         error
	)

	counter := Counter{Count: maxFetchRetries}
	backoffController := BackoffController{BackoffConfig: DefaultBackoffConfig()}

	for counter.Next() {
		logger.Info("checking IP address,", zap.Int("iter", counter.iter))

		if country, ip, err = fetchLocationInfo(logger, proxyParams); err != nil {
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

func fetchLocationInfo(logger *zap.Logger, proxyParams ProxyParams) (country, ip string, err error) {
	const (
		ipCheckerURI   = "https://api.myip.com/"
		requestTimeout = 3 * time.Second
	)

	proxyFunc := GetProxyFunc(proxyParams, "http")

	client := &fasthttp.Client{
		MaxConnDuration:     requestTimeout,
		ReadTimeout:         requestTimeout,
		WriteTimeout:        requestTimeout,
		MaxIdleConnDuration: requestTimeout,
		Dial: func(addr string) (net.Conn, error) {
			return proxyFunc("tcp", addr)
		},
	}

	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI(ipCheckerURI)
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := client.Do(req, resp); err != nil {
		return "", "", err
	}

	ipInfo := struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}{}

	if err := json.Unmarshal(resp.Body(), &ipInfo); err != nil {
		return "", "", err
	}

	return ipInfo.Country, ipInfo.IP, nil
}
