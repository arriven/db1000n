package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	}

	log.Printf("Please open %s", url)
}

// CheckCountry allows to check which country the app is running from
func CheckCountry(countriesToAvoid []string) bool {
	type IPInfo struct {
		Country string `json:"country"`
		IP      string `json:"ip"`
	}

	ipInfo := IPInfo{}
	var ipCheckerURI = "https://api.myip.com/"

	retries := 1
	for ipInfo.IP == "" && retries <= 3 {
		log.Printf("Checking IP address, attempt #%d", retries)
		resp, err := http.Get(ipCheckerURI)
		if err != nil {
			log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
			} else {
				err = json.Unmarshal(body, &ipInfo)

				if err != nil {
					log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
				}
			}
		}

		time.Sleep(1 * time.Second)
		retries += 1
	}

	if retries == 4 {
		return false
	}

	for _, country := range countriesToAvoid {
		if ipInfo.Country == strings.TrimSpace(country) {
			log.Printf("Current country: %s. You might need to enable VPN.", ipInfo.Country)
			openBrowser("https://arriven.github.io/db1000n/vpn/")
			return false
		}
	}

	log.Printf("Current country: %s (%s)", ipInfo.Country, ipInfo.IP)
	return true
}
