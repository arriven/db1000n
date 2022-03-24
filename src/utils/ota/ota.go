// Package ota [allows hot update and reload of the executable]
package ota

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

var (
	// Version is a release version embedded into the app
	Version = "v0.0.1"
	// Repository to check for updates
	Repository = "Arriven/db1000n" // Could be changed via the ldflags
)

// DoAutoUpdate updates the app to the latest version.
func DoAutoUpdate() (updateFound bool, newVersion, changeLog string, err error) {
	v, err := semver.ParseTolerant(Version)
	if err != nil {
		return false, "", "", fmt.Errorf("binary version validation failed: %w", err)
	}

	latest, err := selfupdate.UpdateSelf(v, Repository)

	switch {
	case err != nil:
		return false, "", "", fmt.Errorf("binary update failed: %w", err)
	case latest.Version.Equals(v):
		return false, "", "", nil
	}

	return true, latest.Version.String(), latest.ReleaseNotes, nil
}
