// Package ota [allows hot update and reload of the executable]
package ota

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// DoAutoUpdate updates the app to the latest version.
func DoAutoUpdate() (updateFound bool, newVersion, changeLog string, err error) {
	v, err := semver.ParseTolerant(Version)
	if err != nil {
		err = fmt.Errorf("binary version validation failed: %w", err)

		return
	}

	latest, err := selfupdate.UpdateSelf(v, Repository)
	if err != nil {
		err = fmt.Errorf("binary update failed: %w", err)

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

func appendArgIfNotPresent(osArgs, extraArgs []string) []string {
	osArgsMap := make(map[string]interface{}, len(osArgs))
	for _, osArg := range osArgs {
		osArgsMap[osArg] = nil
	}

	acceptedExtraArgs := make([]string, 0)

	for _, extraArg := range extraArgs {
		if _, isAlreadyOSArg := osArgsMap[extraArg]; !isAlreadyOSArg {
			acceptedExtraArgs = append(acceptedExtraArgs, extraArg)
		}
	}

	return append(osArgs, acceptedExtraArgs...)
}
