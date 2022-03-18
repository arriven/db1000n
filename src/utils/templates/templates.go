// Package templates [provides utility functions to enable templating in app configuration]
package templates

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/corpix/uarand"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(b, &urls)
	if err != nil {
		// try to parse response body as plain text with newline delimiter
		urls = strings.Split(string(b), "\n")
		if len(urls) == 0 {
			return nil
		}
	}

	return urls
}

func getURLContent(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
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

// ContextKey used to work with context and not trigger linter
type ContextKey string

func ctxKey(key string) ContextKey {
	return ContextKey(key)
}

func cookieString(cookies map[string]string) string {
	s := ""
	for key, value := range cookies {
		s = fmt.Sprintf("%s %s=%s;", s, key, value)
	}

	return strings.Trim(strings.TrimSpace(s), ";")
}

// Parse a template
func Parse(input string) (*template.Template, error) {
	// TODO: consider adding ability to populate custom data
	return template.New("tpl").Funcs(template.FuncMap{
		"random_uuid":          randomUUID,
		"random_int_n":         rand.Intn,
		"random_int":           rand.Int,
		"random_payload":       RandomPayload,
		"random_ip":            RandomIP,
		"random_port":          RandomPort,
		"random_mac_addr":      RandomMacAddr,
		"random_user_agent":    uarand.GetRandom,
		"local_ip":             LocalIPV4,
		"local_ipv4":           LocalIPV4,
		"local_ipv6":           LocalIPV6,
		"local_mac_addr":       LocalMacAddres,
		"resolve_host":         ResolveHostIPV4,
		"resolve_host_ipv4":    ResolveHostIPV4,
		"resolve_host_ipv6":    ResolveHostIPV6,
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
		"ctx_key":              ctxKey,
		"split":                strings.Split,
		"cookie_string":        cookieString,
	}).Parse(strings.ReplaceAll(input, "\\", ""))
}

// Execute template, returns empty string in case of errors
func Execute(logger *zap.Logger, tpl *template.Template, data interface{}) string {
	var res strings.Builder
	if err := tpl.Execute(&res, data); err != nil {
		logger.Error("error executing template", zap.Error(err))

		return ""
	}

	return res.String()
}

// ParseAndExecute template, returns input string in case of errors. Expensive operation.
func ParseAndExecute(logger *zap.Logger, input string, data interface{}) string {
	tpl, err := Parse(input)
	if err != nil {
		logger.Error("error parsing template", zap.Error(err))

		return input
	}

	var output strings.Builder
	if err = tpl.Execute(&output, data); err != nil {
		logger.Error("error executing template", zap.Error(err))

		return input
	}

	return output.String()
}

// ParseAndExecuteMapStruct is like ParseAndExecute but takes mapstructure as input
func ParseAndExecuteMapStruct(logger *zap.Logger, input map[string]interface{}, data interface{}) map[string]interface{} {
	tpl, err := ParseMapStruct(input)
	if err != nil {
		logger.Error("error parsing template", zap.Error(err))

		return input
	}

	return tpl.Execute(logger, data)
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
func (tpl *MapStruct) Execute(logger *zap.Logger, data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range tpl.tpl {
		switch v := value.(type) {
		case *template.Template:
			result[key] = Execute(logger, v, data)
		case *MapStruct:
			result[key] = v.Execute(logger, data)
		default:
			result[key] = v
		}
	}

	return result
}
