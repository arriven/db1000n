// Package ota [allows hot update and reload of the executable]
package ota

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// DoAutoUpdate updates the app to the latest version.
func DoAutoUpdate() (updateFound bool, newVersion, changeLog string, err error) {
	v, err := semver.ParseTolerant(Version)
	if err != nil {
		err = errors.Wrap(err, "binary version validation failed")
		return
	}

	latest, err := selfupdate.UpdateSelf(v, Repository)
	if err != nil {
		err = errors.Wrap(err, "binary update failed")
		return
	}

	if !latest.Version.Equals(v) {
		updateFound = true
		newVersion = latest.Version.String()
		changeLog = latest.ReleaseNotes
	}

	return
}

func MockAutoUpdate(shouldUpdateBeFound bool) (updateFound bool, newVersion, changeLog string, err error) {
	updateFound = shouldUpdateBeFound
	if updateFound {
		newVersion = "brand-new-version"
		changeLog = "something-was-changed"
		err = nil
	}

	return
}
