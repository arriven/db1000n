package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/Arriven/db1000n/src/logs"
)

func openBrowser(url string, l *logs.Logger) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}
	l.Info("Please open %s", url)
}

// CheckCountry allows to check which country the app is running from
func CheckCountry(l *logs.Logger, countriesToAvoid []string) {
	type IPInfo struct {
		Country string `json:"country"`
	}

	resp, err := http.Get("https://api.myip.com/")
	if err != nil {
		l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}
	ipInfo := IPInfo{}
	err = json.Unmarshal(body, &ipInfo)
	if err != nil {
		l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}

	for _, country := range countriesToAvoid {
		if ipInfo.Country == country {
			l.Warning("Current country: %s. You might need to enable VPN.", ipInfo.Country)
			// TODO add correct URL
			//openBrowser("https://example.com/", l)
			return
		}
	}
	l.Info("Current country: %s", ipInfo.Country)
}
