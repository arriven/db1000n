package countrychecker

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"strings"
	"time"

	"github.com/Arriven/db1000n/src/core/http"
	"github.com/Arriven/db1000n/src/job"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
	"github.com/corpix/uarand"
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

	flag.StringVar(&res.countryBlackListCSV, "country-list", utils.GetEnvStringDefault("COUNTRY_LIST", "Ukraine"), "comma-separated list of countries")
	flag.BoolVar(&res.strictCountryCheck, "strict-country-check", utils.GetEnvBoolDefault("STRICT_COUNTRY_CHECK", false),
		"enable strict country check; will also exit if IP can't be determined")
	flag.IntVar(&res.maxRetries, "country-check-retries", utils.GetEnvIntDefault("COUNTRY_CHECK_RETRIES", maxFetchRetries),
		"how much retries should be made when checking the country")

	return &res
}

// CheckCountryOrFail checks the country of client origin by IP and exits the program if it is in the blacklist.
func CheckCountryOrFail(logger *zap.Logger, cfg *CountryCheckerConfig, globalConfig *job.GlobalConfig) string {
	countriesToAvoid := strings.Split(cfg.countryBlackListCSV, ",")
	isCountryAllowed, country := CheckCountry(logger, countriesToAvoid, cfg.strictCountryCheck, globalConfig, cfg.maxRetries)
	if !isCountryAllowed {
		logger.Fatal("not an allowed country, exiting", zap.String("country", country))
	}

	return country
}

// CheckCountry checks which country the app is running from and whether it is in the blacklist.
func CheckCountry(logger *zap.Logger, countriesToAvoid []string, strictCountryCheck bool, globalConfig *job.GlobalConfig, maxFetchRetries int) (bool, string) {
	var (
		country, ip string
		err         error
	)

	counter := utils.Counter{Count: maxFetchRetries}
	backoffController := utils.BackoffController{BackoffConfig: utils.DefaultBackoffConfig()}

	for counter.Next() {
		logger.Info("checking IP address,", zap.Int("iter", counter.Get()))

		if country, ip, err = fetchLocationInfoFastHTTP(logger, globalConfig); err != nil {
			logger.Warn("error fetching location info", zap.Error(err))
			utils.Sleep(context.Background(), backoffController.Increment().GetTimeout())
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

func fetchLocationInfoFastHTTP(logger *zap.Logger, global *job.GlobalConfig) (country, ip string, err error) {
	ipCheckerURI := "https://api.myip.com/"
	requestTimeout := 3 * time.Second
	maxIdleConns := 10

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	clientConfig := &http.ClientConfig{
		StaticHost:      nil,
		TLSClientConfig: nil,
		Timeout:         &requestTimeout,
		ReadTimeout:     &requestTimeout,
		WriteTimeout:    &requestTimeout,
		IdleTimeout:     &requestTimeout,
		MaxIdleConns:    &maxIdleConns,
		ProxyURLs:       "",
		LocalAddr:       "",
		Interface:       "",
	}
	if global.ProxyURLs != "" {
		clientConfig.ProxyURLs = templates.ParseAndExecute(logger, global.ProxyURLs, ctx)
	}

	if global.LocalAddr != "" {
		clientConfig.LocalAddr = templates.ParseAndExecute(logger, global.LocalAddr, ctx)
	}

	if global.Interface != "" {
		clientConfig.Interface = templates.ParseAndExecute(logger, global.Interface, ctx)
	}

	client := http.NewClient(ctx, *clientConfig, logger)

	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	logger.Info("single http IP request", zap.String("target", ipCheckerURI))

	req.SetRequestURI(ipCheckerURI)
	req.Header.SetMethod(fasthttp.MethodGet)

	// Add random user agent and configured headers
	req.Header.Set("user-agent", uarand.GetRandom())

	if err := client.Do(req, resp); err != nil {
		return "", "", err
	}

	ipInfo := struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}{}

	if err := json.NewDecoder(bytes.NewReader(resp.Body())).Decode(&ipInfo); err != nil {
		return "", "", err
	}
	return ipInfo.Country, ipInfo.IP, nil
}
