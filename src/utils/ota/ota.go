// Package ota [allows hot update and reload of the executable]
package ota

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// DoSelfUpdate updates the app to the latest version
func DoSelfUpdate() {
	v, err := semver.ParseTolerant(Version)
	if err != nil {
		log.Println("Binary version validation failed:", err)
		return
	}

	latest, err := selfupdate.UpdateSelf(v, Repository)
	if err != nil {
		log.Println("Binary update failed:", err)
		return
	}

	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up-to-date.
		log.Println("Current binary is the latest version", Version)
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
	}
}

// ConfirmAndSelfUpdate ask for user confirmation to do a self-update
func ConfirmAndSelfUpdate() {
	latest, found, err := selfupdate.DetectLatest(Repository)
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	if v := semver.MustParse(Version); !found || latest.Version.LTE(v) {
		log.Println("Current version is the latest")
		return
	}

	fmt.Print("Do you want to update to", latest.Version, "? (y/n): ") //nolint:forbidigo // Here we actually write to console and expect user input
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		log.Println("Invalid input")
		return
	}
	if input == "n\n" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Println("Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version)
}
