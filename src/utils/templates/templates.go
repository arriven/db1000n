package templates

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"text/template"

	"github.com/google/uuid"

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

func mod(lhs, rhs uint32) uint32 {
	return lhs % rhs
}

// Parse a template
func Parse(input string) (*template.Template, error) {
	// TODO: consider adding ability to populate custom data
	return template.New("tpl").Funcs(template.FuncMap{
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
		"mod":                  mod,
	}).Parse(strings.Replace(input, "\\", "", -1))
}

// Execute template, returns empty string in case of errors
func Execute(tpl *template.Template, data interface{}) string {
	var res strings.Builder
	if err := tpl.Execute(&res, data); err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return res.String()
}

// ParseAndExecute template, returns input string in case of errors. Expensive operation.
func ParseAndExecute(input string, data interface{}) string {
	tpl, err := Parse(input)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return input
	}

	var output strings.Builder
	if err = tpl.Execute(&output, data); err != nil {
		log.Printf("Error executing template: %v", err)
		return input
	}

	return output.String()
}

// ParseAndExecuteMapStruct is like ParseAndExecute but takes mapstructure as input
func ParseAndExecuteMapStruct(input map[string]interface{}, data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range input {
		switch v := value.(type) {
		case string:
			result[key] = ParseAndExecute(v, data)
		case map[string]interface{}:
			result[key] = ParseAndExecuteMapStruct(v, data)
		default:
			result[key] = v
		}
	}
	return result
}
