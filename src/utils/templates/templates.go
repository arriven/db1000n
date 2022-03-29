// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package templates [provides utility functions to enable templating in app configuration]
package templates

import (
	"context"
	"encoding/base64"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/corpix/uarand"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

func randomChar(from string) byte {
	if len(from) == 0 {
		return 0
	}

	return from[rand.Intn(len(from))] //nolint:gosec // Cryptographically secure random not required
}

func randomString(n int, from string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomChar(from)
	}

	return string(b)
}

func randomAlpha(n int) string {
	return randomString(n, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func randomAplhaNum(n int) string {
	return randomString(n, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}

func mod(lhs, rhs int) int {
	return lhs % rhs
}

func add(lhs, rhs int) int {
	return lhs + rhs
}

// ContextKey used to work with context and not trigger linter
type ContextKey string

func ctxKey(key string) ContextKey {
	return ContextKey(key)
}

// Parse a template
func Parse(input string) (*template.Template, error) {
	// TODO: consider adding ability to populate custom data
	return template.New("tpl").Funcs(template.FuncMap{
		"random_uuid":         randomUUID,
		"random_char":         randomChar,
		"random_string":       randomString,
		"random_alpha":        randomAlpha,
		"random_alphanum":     randomAplhaNum,
		"random_int_n":        rand.Intn,
		"random_int":          rand.Int,
		"random_payload":      RandomPayload,
		"random_payload_byte": RandomPayloadByte,
		"random_ip":           RandomIP,
		"random_port":         RandomPort,
		"random_mac_addr":     RandomMacAddr,
		"random_user_agent":   uarand.GetRandom,
		"local_ip":            LocalIPV4,
		"local_ipv4":          LocalIPV4,
		"local_ipv6":          LocalIPV6,
		"local_mac_addr":      LocalMacAddres,
		"resolve_host":        ResolveHostIPV4,
		"resolve_host_ipv4":   ResolveHostIPV4,
		"resolve_host_ipv6":   ResolveHostIPV6,
		"base64_encode":       base64.StdEncoding.EncodeToString,
		"base64_decode":       base64.StdEncoding.DecodeString,
		"to_yaml":             toYAML,
		"from_yaml":           fromYAML,
		"from_yaml_array":     fromYAMLArray,
		"to_json":             toJSON,
		"from_json":           fromJSON,
		"from_json_array":     fromJSONArray,
		"from_string_array":   fromStringArray,
		"join":                strings.Join,
		"split":               strings.Split,
		"get_url":             getURLContent,
		"mod":                 mod,
		"add":                 add,
		"ctx_key":             ctxKey,
		"cookie_string":       cookieString,
	}).Parse(input)
}

// Execute template, returns empty string in case of errors
func Execute(logger *zap.Logger, tpl *template.Template, data interface{}) string {
	var res strings.Builder
	if err := tpl.Execute(&res, data); err != nil {
		logger.Warn("error executing template", zap.Error(err))

		return ""
	}

	return res.String()
}

// ParseAndExecute template, returns input string in case of errors. Expensive operation.
func ParseAndExecute(logger *zap.Logger, input string, data interface{}) string {
	tpl, err := Parse(input)
	if err != nil {
		logger.Warn("error parsing template", zap.Error(err))

		return input
	}

	var output strings.Builder
	if err = tpl.Execute(&output, data); err != nil {
		logger.Warn("error executing template", zap.Error(err))

		return input
	}

	return output.String()
}

// ParseAndExecuteMapStruct is like ParseAndExecute but takes mapstructure as input
func ParseAndExecuteMapStruct(logger *zap.Logger, input map[string]interface{}, data interface{}) map[string]interface{} {
	tpl, err := ParseMapStruct(input)
	if err != nil {
		logger.Warn("error parsing template", zap.Error(err))

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
