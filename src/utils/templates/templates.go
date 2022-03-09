// Package templates [provides utility functions to enable templating in app configuration]
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
	"gopkg.in/yaml.v3"

	"github.com/Arriven/db1000n/src/packetgen"
)

var proxiesURL = "https://raw.githubusercontent.com/Arriven/db1000n/main/proxylist.json"

func getProxylistURL() string {
	return proxiesURL
}

// SetProxiesURL is used to override the default proxylist url
func SetProxiesURL(url string) {
	proxiesURL = url
}

func getProxylist() (urls []string) {
	return getProxylistByURL(getProxylistURL())
}

func getProxylistByURL(url string) (urls []string) {
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
		"resolve_host":         packetgen.ResolveHost,
		"base64_encode":        base64.StdEncoding.EncodeToString,
		"base64_decode":        base64.StdEncoding.DecodeString,
		"json_encode":          json.Marshal,
		"json_decode":          json.Unmarshal,
		"yaml_encode":          yaml.Marshal,
		"yaml_decode":          yaml.Unmarshal,
		"join":                 strings.Join,
		"get_url":              getURLContent,
		"proxylist_url":        getProxylistURL,
		"get_proxylist":        getProxylist,
		"get_proxylist_by_url": getProxylistByURL,
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
	tpl, err := ParseMapStruct(input)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return input
	}
	return tpl.Execute(data)
}

// MapStruct is a helper structure to parse configs in a format accepted by mapstructure package
type MapStruct struct {
	tpl map[string]interface{}
}

// ParseMapStruct is like Parse but takes mapstructure as input
func ParseMapStruct(input map[string]interface{}) (*MapStruct, error) {
	result := make(map[string]interface{})
	for key, value := range input {
		switch v := value.(type) {
		case string:
			tpl, err := Parse(v)
			if err != nil {
				return nil, err
			}
			result[key] = tpl
		case map[string]interface{}:
			tpl, err := ParseMapStruct(v)
			if err != nil {
				return nil, err
			}
			result[key] = tpl
		default:
			result[key] = v
		}
	}
	return &MapStruct{tpl: result}, nil
}

// Execute same as regular Execute
func (tpl *MapStruct) Execute(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range tpl.tpl {
		switch v := value.(type) {
		case *template.Template:
			result[key] = Execute(v, data)
		case *MapStruct:
			result[key] = v.Execute(data)
		default:
			result[key] = v
		}
	}
	return result
}
