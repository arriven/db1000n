package config

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/Arriven/db1000n/src/utils/ota"
)

//nolint: paralleltest // No need to test this in parallel :p
func TestGetDefaultConfigURL(t *testing.T) {
	t.Run("Case when the version is not embedded", func(tt *testing.T) {
		ota.Version = ota.InitialVersion

		resolvedDefaultConfigURL := GetDefaultConfigURL()

		t.Logf("Resolved version: %s", resolvedDefaultConfigURL)

		if resolvedDefaultConfigURL != FallbackConfigURL {
			tt.Errorf("unexpected default URL:\nexp:%s\ngot:%s",
				FallbackConfigURL, resolvedDefaultConfigURL)

			return
		}

		if _, err := url.Parse(resolvedDefaultConfigURL); err != nil {
			tt.Errorf("unexpected error validation the default URL: %s", err)

			return
		}
	})

	t.Run("Case when the embedded version has invalid value", func(tt *testing.T) {
		ota.Version = "develop"

		resolvedDefaultConfigURL := GetDefaultConfigURL()

		t.Logf("Resolved version: %s", resolvedDefaultConfigURL)

		if resolvedDefaultConfigURL != FallbackConfigURL {
			tt.Errorf("unexpected default URL:\nexp:%s\ngot:%s",
				FallbackConfigURL, resolvedDefaultConfigURL)

			return
		}

		if _, err := url.Parse(resolvedDefaultConfigURL); err != nil {
			tt.Errorf("unexpected error validation the default URL: %s", err)

			return
		}
	})

	t.Run("Case the embedded version is valid and different from the initial", func(tt *testing.T) {
		var (
			majorVer = uint(0)
			minorVer = uint(8)
			patchVer = uint(0)
		)

		ota.Version = fmt.Sprintf("v%d.%d.%d", majorVer, minorVer, patchVer)

		resolvedDefaultConfigURL := GetDefaultConfigURL()

		t.Logf("Resolved version: %s", resolvedDefaultConfigURL)

		if resolvedDefaultConfigURL != buildDefaultConfigURLForVersion(majorVer, minorVer, patchVer) {
			tt.Errorf("unexpected default URL:\nexp:%s\ngot:%s",
				FallbackConfigURL, resolvedDefaultConfigURL)
		}

		if _, err := url.Parse(resolvedDefaultConfigURL); err != nil {
			tt.Errorf("unexpected error validation the default URL: %s", err)

			return
		}
	})
}
