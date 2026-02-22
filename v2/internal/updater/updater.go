package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CurrentVersion is the app version. Used everywhere instead of hardcoded strings.
const CurrentVersion = "v2.0.2"

// UpdateInfo describes a single available update.
type UpdateInfo struct {
	Version    string
	ReleaseURL string
	IsBeta     bool
}

// Result holds the outcome of an update check.
type Result struct {
	StableUpdate *UpdateInfo
	BetaUpdate   *UpdateInfo
}

type ghRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	HTMLURL    string `json:"html_url"`
}

// CheckAsync runs an update check in a background goroutine.
// The callback is only called when at least one update is found.
// Errors (network, parse, etc.) are silently ignored.
func CheckAsync(cb func(Result)) {
	go func() {
		result, err := check()
		if err != nil {
			return
		}
		if result.StableUpdate != nil || result.BetaUpdate != nil {
			cb(result)
		}
	}()
}

func check() (Result, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET",
		"https://api.github.com/repos/miwidot/tarkovaccountswitcher/releases", nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("github api: %d", resp.StatusCode)
	}

	var releases []ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return Result{}, err
	}

	curMajor, curMinor, curPatch, ok := parseSemver(CurrentVersion)
	if !ok {
		return Result{}, fmt.Errorf("cannot parse current version")
	}

	var result Result

	for _, rel := range releases {
		if rel.Draft {
			continue
		}

		major, minor, patch, ok := parseSemver(rel.TagName)
		if !ok {
			continue
		}

		if !isNewer(major, minor, patch, curMajor, curMinor, curPatch) {
			continue
		}

		info := &UpdateInfo{
			Version:    rel.TagName,
			ReleaseURL: rel.HTMLURL,
			IsBeta:     rel.Prerelease,
		}

		if rel.Prerelease {
			if result.BetaUpdate == nil {
				result.BetaUpdate = info
			} else {
				bm, bn, bp := mustParseSemver(result.BetaUpdate.Version)
				if isNewer(major, minor, patch, bm, bn, bp) {
					result.BetaUpdate = info
				}
			}
		} else {
			if result.StableUpdate == nil {
				result.StableUpdate = info
			} else {
				sm, sn, sp := mustParseSemver(result.StableUpdate.Version)
				if isNewer(major, minor, patch, sm, sn, sp) {
					result.StableUpdate = info
				}
			}
		}
	}

	return result, nil
}

// parseSemver parses "v2.1.0" or "2.1.0" into (2,1,0,true).
func parseSemver(s string) (int, int, int, bool) {
	s = strings.TrimPrefix(s, "v")
	// Strip any pre-release suffix like "-beta.1"
	if idx := strings.IndexByte(s, '-'); idx != -1 {
		s = s[:idx]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return major, minor, patch, true
}

func mustParseSemver(s string) (int, int, int) {
	major, minor, patch, _ := parseSemver(s)
	return major, minor, patch
}

// isNewer returns true if a.b.c > x.y.z.
func isNewer(a, b, c, x, y, z int) bool {
	if a != x {
		return a > x
	}
	if b != y {
		return b > y
	}
	return c > z
}
