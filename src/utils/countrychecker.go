package utils

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type CountryCheckerConfig struct {
	countryBlackListCSV string
	strict              bool
	interval            time.Duration
	maxRetries          int
}

// NewGlobalConfigWithFlags returns a GlobalConfig initialized with command line flags.
func NewCountryCheckerConfigWithFlags() *CountryCheckerConfig {
	const maxFetchRetries = 3

	var res CountryCheckerConfig

	flag.StringVar(&res.countryBlackListCSV, "country-list", GetEnvStringDefault("COUNTRY_LIST", "Ukraine"), "comma-separated list of countries")
	flag.BoolVar(&res.strict, "strict-country-check", GetEnvBoolDefault("STRICT_COUNTRY_CHECK", false),
		"enable strict country check; will also exit if IP can't be determined")
	flag.IntVar(&res.maxRetries, "country-check-retries", GetEnvIntDefault("COUNTRY_CHECK_RETRIES", maxFetchRetries),
		"how much retries should be made when checking the country")
	flag.DurationVar(&res.interval, "country-check-interval", GetEnvDurationDefault("COUNTRY_CHECK_INTERVAL", 0),
		"run country check in background with a regular interval")

	return &res
}

// CheckCountryOrFail checks the country of client origin by IP and exits the program if it is in the blacklist.
func CheckCountryOrFail(ctx context.Context, logger *zap.Logger, cfg *CountryCheckerConfig, proxyParams ProxyParams) string {
	if cfg.interval != 0 {
		go func() {
			ticker := time.NewTicker(cfg.interval)

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_ = ckeckCountryOnce(logger, cfg, proxyParams)
				}
			}
		}()
	}

	return ckeckCountryOnce(logger, cfg, proxyParams)
}

func ckeckCountryOnce(logger *zap.Logger, cfg *CountryCheckerConfig, proxyParams ProxyParams) string {
	country, ip, err := getCountry(logger, proxyParams, cfg.maxRetries)
	if err != nil {
		if cfg.strict {
			logger.Fatal("country strict check failed", zap.Error(err))
		}

		return ""
	}

	logger.Info("location info", zap.String("country", country), zap.String("ip", ip))

	if strings.Contains(cfg.countryBlackListCSV, country) {
		logger.Warn("you might need to enable VPN.")

		if cfg.strict {
			logger.Fatal("country strict check failed", zap.String("country", country))
		}
	}

	return country
}

func getCountry(logger *zap.Logger, proxyParams ProxyParams, maxFetchRetries int) (country, ip string, err error) {
	counter := Counter{Count: maxFetchRetries}
	backoffController := BackoffController{BackoffConfig: DefaultBackoffConfig()}

	for counter.Next() {
		logger.Info("checking IP address,", zap.Int("iter", counter.iter))

		if country, ip, err = fetchLocationInfo(logger, proxyParams); err != nil {
			logger.Warn("error fetching location info", zap.Error(err))
			Sleep(context.Background(), backoffController.Increment().GetTimeout())
		} else {
			return
		}
	}

	return "", "", fmt.Errorf("couldn't get location info in %d tries", maxFetchRetries)
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
