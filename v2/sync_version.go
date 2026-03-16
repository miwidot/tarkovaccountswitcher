//go:build ignore

// sync_version reads CurrentVersion from internal/updater/updater.go
// and updates productVersion in wails.json to match.
// Run before building: go run sync_version.go && wails build
package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Read version from updater.go
	src, err := os.ReadFile("internal/updater/updater.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading updater.go: %v\n", err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`CurrentVersion\s*=\s*"v([^"]+)"`)
	m := re.FindSubmatch(src)
	if m == nil {
		fmt.Fprintln(os.Stderr, "Error: CurrentVersion not found in updater.go")
		os.Exit(1)
	}
	version := string(m[1]) // e.g. "2.0.4"

	// Read wails.json and replace productVersion in-place (preserves key order)
	data, err := os.ReadFile("wails.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading wails.json: %v\n", err)
		os.Exit(1)
	}

	pvRe := regexp.MustCompile(`("productVersion"\s*:\s*")([^"]*)(")`)
	loc := pvRe.FindSubmatchIndex(data)
	if loc == nil {
		fmt.Fprintln(os.Stderr, "Error: productVersion not found in wails.json")
		os.Exit(1)
	}

	oldVersion := string(data[loc[4]:loc[5]])
	if oldVersion == version {
		fmt.Printf("Version already in sync: %s\n", version)
		return
	}

	updated := pvRe.ReplaceAll(data, []byte(`${1}`+version+`${3}`))

	if err := os.WriteFile("wails.json", updated, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing wails.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated wails.json: %s -> %s\n", oldVersion, version)
}
