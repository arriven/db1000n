package templates

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"text/template"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/packetgen"
)

var proxiesURL = "https://raw.githubusercontent.com/Arriven/db1000n/main/proxylist.json"

func getProxylistURL() string {
	return proxiesURL
}

func SetProxiesUrl(url string) {
	proxiesURL = url
}

func getProxylist() (urls []string) {
	return getProxylistByUrl(getProxylistURL())
}

func getProxylistByUrl(url string) (urls []string) {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&urls); err != nil {
		return nil
	}
	return urls
}
func getURLContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func randomUUID() string {
	return uuid.New().String()
}

func Execute(input string) string {
	funcMap := template.FuncMap{
		"random_uuid":          randomUUID,
		"random_int_n":         rand.Intn,
		"random_int":           rand.Int,
		"random_payload":       packetgen.RandomPayload,
		"random_ip":            packetgen.RandomIP,
		"random_port":          packetgen.RandomPort,
		"random_mac_addr":      packetgen.RandomMacAddr,
		"local_ip":             packetgen.LocalIP,
		"local_mac_addr":       packetgen.LocalMacAddres,
		"base64_encode":        base64.StdEncoding.EncodeToString,
		"base64_decode":        base64.StdEncoding.DecodeString,
		"json_encode":          json.Marshal,
		"json_decode":          json.Unmarshal,
		"get_url":              getURLContent,
		"proxylist_url":        getProxylistURL,
		"get_proxylist":        getProxylist,
		"get_proxylist_by_url": getProxylistByUrl,
	}

	// TODO: consider adding ability to populate custom data
	input = strings.Replace(input, "\\", "", -1)
	tmpl, err := template.New("test").Funcs(funcMap).Parse(input)
	if err != nil {
		logs.Default.Warning("error parsing template: %v", err)
		return input
	}

	var output strings.Builder
	if err = tmpl.Execute(&output, nil); err != nil {
		logs.Default.Warning("error executing template: %v", err)
		return input
	}

	return output.String()
}
