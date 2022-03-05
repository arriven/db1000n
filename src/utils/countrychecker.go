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
func CheckCountry(l *logs.Logger) {
	type IPInfo struct {
		Country string `json:"country"`
	}

	resp, err := http.Get("https://api.myip.com/")
	if err != nil {
		l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		} else {
			ipInfo := IPInfo{}
			err = json.Unmarshal(body, &ipInfo)
			if err != nil {
				l.Warning("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
			} else {
				if ipInfo.Country == "Ukraine" {
					l.Error("You currently have Ukrainian IP adsress. You need to enable VPN.")
					// TODO add correct URL
					//openBrowser("https://example.com/", l)
				} else {
					l.Info("Current country: %s", ipInfo.Country)
				}
			}

		}
	}
}
