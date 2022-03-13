package utils

import (
	"context"
	"encoding/json"
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
func CheckCountry(ctx context.Context, countriesToAvoid []string) {
	const ipCheckerURI = "https://api.myip.com/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipCheckerURI, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Can't check country. Please manually check that VPN is enabled or that you have non-Ukrainian IP address.")

		return
	}

	defer resp.Body.Close()

	var ipInfo struct {
		Country string `json:"country"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		log.Println("Can't check country. Please manually check that VPN is enabled or that you have non-Ukrainian IP address.")

		return
	}

	for _, country := range countriesToAvoid {
		if ipInfo.Country == country {
			log.Printf("Current country: %s. You might need to enable VPN.", ipInfo.Country)
			openBrowser("https://arriven.github.io/db1000n/vpn/")

			return
		}
	}

	log.Printf("Current country: %s", ipInfo.Country)
}
