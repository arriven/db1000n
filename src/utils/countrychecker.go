package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
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
func CheckCountry() {
	type IPInfo struct {
		Country string `json:"country"`
	}

	resp, err := http.Get("https://api.myip.com/")
	if err != nil {
		log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}

	ipInfo := IPInfo{}
	if err = json.Unmarshal(body, &ipInfo); err != nil {
		log.Println("Can't check users country. Please manually check that VPN is enabled or that you have non Ukrainian IP address.")
		return
	}

	if ipInfo.Country == "Ukraine" {
		log.Println("You currently have Ukrainian IP adsress. You need to enable VPN.")
		// TODO add correct URL
		// openBrowser("https://example.com/", l)
		return
	}

	log.Printf("Current country: %s", ipInfo.Country)
}
