//go:build ignore

// sync_version reads CurrentVersion from internal/updater/updater.go and keeps
// every version-bearing build input in sync with it:
//   - wails.json          productVersion (PE/app version Wails reports)
//   - build/windows/info.json  the version resource Wails embeds into the .exe
//
// The Wails default info.json template only fills fixed.file_version, leaving
// the binary's ProductVersion at 0.0.0.0; writing our own info.json sets BOTH
// fixed.file_version and fixed.product_version so Explorer shows 2.0.7.0 for
// both. build/ is gitignored, so this regenerates correctly on every machine.
//
// Run before building: go run sync_version.go && wails build
package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	version := string(m[1]) // e.g. "2.0.7"

	syncWailsJSON(version)
	writeWindowsInfoJSON(version)
}

// syncWailsJSON rewrites productVersion in wails.json in-place (preserves key order).
func syncWailsJSON(version string) {
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
		fmt.Printf("wails.json already in sync: %s\n", version)
		return
	}

	updated := pvRe.ReplaceAll(data, []byte(`${1}`+version+`${3}`))
	if err := os.WriteFile("wails.json", updated, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing wails.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated wails.json: %s -> %s\n", oldVersion, version)
}

// writeWindowsInfoJSON generates build/windows/info.json with both the file and
// product version set, so Wails embeds a complete VS_FIXEDFILEINFO.
func writeWindowsInfoJSON(version string) {
	path := filepath.Join("build", "windows", "info.json")
	content := fmt.Sprintf(`{
  "fixed": {
    "file_version": "%[1]s.0",
    "product_version": "%[1]s.0"
  },
  "info": {
    "0409": {
      "ProductVersion": "%[1]s.0",
      "FileVersion": "%[1]s.0",
      "InternalName": "Tarkov Account Switcher",
      "OriginalFilename": "Tarkov Account Switcher.exe",
      "CompanyName": "Tarkov-Stammtisch.de",
      "FileDescription": "Tarkov Account Switcher",
      "LegalCopyright": "Tarkov-Stammtisch.de",
      "ProductName": "Tarkov Account Switcher",
      "Comments": "Manage multiple Escape from Tarkov accounts"
    }
  }
}
`, version)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating build/windows dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s: file/product version %s.0\n", path, version)
}
