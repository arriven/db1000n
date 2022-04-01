package templates

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Arriven/db1000n/src/utils"
)

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v any) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}

	return strings.TrimSuffix(string(data), "\n")
}

// fromYAML converts a YAML document into a map[string]any.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromYAML(str string) map[string]any {
	m := map[string]any{}
	if err := utils.Unmarshal([]byte(str), &m, "yaml"); err != nil {
		m["Error"] = err.Error()
	}

	return m
}

// fromYAMLArray converts a YAML array into a []any.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromYAMLArray(str string) []any {
	a := []any{}
	if err := utils.Unmarshal([]byte(str), &a, "yaml"); err != nil {
		a = []any{err.Error()}
	}

	return a
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}

	return string(data)
}

// fromJSON converts a JSON document into a map[string]any.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromJSON(str string) map[string]any {
	m := make(map[string]any)
	if err := utils.Unmarshal([]byte(str), &m, "json"); err != nil {
		m["Error"] = err.Error()
	}

	return m
}

// fromJSONArray converts a JSON array into a []any.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromJSONArray(str string) []any {
	a := []any{}
	if err := utils.Unmarshal([]byte(str), &a, "json"); err != nil {
		a = []any{err.Error()}
	}

	return a
}

func fromStringArray(str string) []string {
	a := []string{}
	if err := utils.Unmarshal([]byte(str), &a, "yaml"); err != nil {
		a = []string{err.Error()}
	}

	return a
}
