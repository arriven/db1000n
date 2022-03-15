package ota

import (
	"fmt"

	"github.com/blang/semver"
)

var (
	// Version is a release version embedded into the app
	Version = InitialVersion
	// Repository to check for updates
	Repository = "Arriven/db1000n" // Could be changed via the ldflags
)

const (
	InitialVersion = "v0.0.1"
)

func IsVersionEmbeddedToBinary() (bool, error) {
	resolvedVersionP, err := semver.ParseTolerant(Version)
	if err != nil {
		return false, fmt.Errorf("resolved binary version validation failed: %w", err)
	}

	initialVersionP, err := semver.ParseTolerant(InitialVersion)
	if err != nil {
		return false, fmt.Errorf("initial binary version validation failed: %w", err)
	}

	return !resolvedVersionP.EQ(initialVersionP), nil
}

func ResolveVersion() (major, minor, patch uint, err error) {
	resolvedVersionP, err := semver.ParseTolerant(Version)
	if err != nil {
		err = fmt.Errorf("binary version validation failed: %w", err)

		return
	}

	major = uint(resolvedVersionP.Major)
	minor = uint(resolvedVersionP.Minor)
	patch = uint(resolvedVersionP.Patch)

	return
}
