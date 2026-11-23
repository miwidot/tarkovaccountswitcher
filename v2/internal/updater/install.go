package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

const (
	// Updates are only ever downloaded from this repository's releases.
	downloadPrefix = "https://github.com/miwidot/tarkovaccountswitcher/releases/download/"

	// maxUpdateSize bounds the download; the exe is about 15 MB.
	maxUpdateSize = 200 << 20

	// afterUpdateFlag is passed to the new exe together with the PID of the
	// instance it replaces.
	afterUpdateFlag = "--after-update"
)

// Install downloads the release with the given tag, verifies it and puts it in
// place of the running exe. The running exe is kept as "<exe>.old" until the
// next start. Call Restart afterwards.
func Install(tag string) error {
	asset, err := findAsset(tag)
	if err != nil {
		return err
	}
	exe, err := executablePath()
	if err != nil {
		return err
	}

	newExe := exe + ".new"
	if err := download(asset, newExe); err != nil {
		os.Remove(newExe)
		return err
	}
	// A valid signature alone isn't enough: it must be ours.
	signer, err := verifySignature(newExe)
	if err == nil && signer != signerName {
		err = fmt.Errorf("update is signed by %q, expected %q", signer, signerName)
	}
	if err != nil {
		os.Remove(newExe)
		return err
	}

	// Windows allows renaming a running exe, but not overwriting it.
	oldExe := exe + ".old"
	os.Remove(oldExe)
	if err := os.Rename(exe, oldExe); err != nil {
		os.Remove(newExe)
		return err
	}
	if err := os.Rename(newExe, exe); err != nil {
		os.Rename(oldExe, exe)
		os.Remove(newExe)
		return err
	}
	return nil
}

// Restart starts the exe that Install put in place. The caller must quit
// afterwards; the new instance waits for that (see FinishUpdate).
func Restart() error {
	exe, err := executablePath()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, afterUpdateFlag, strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// FinishUpdate runs at startup. Started by Restart, it waits for the previous
// instance to exit, which otherwise still holds the single-instance lock, and
// then removes the exe that instance ran from.
func FinishUpdate() {
	for i, arg := range os.Args {
		if arg != afterUpdateFlag || i+1 >= len(os.Args) {
			continue
		}
		pid, err := strconv.Atoi(os.Args[i+1])
		if err != nil {
			break
		}
		h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
		if err != nil {
			break // already gone
		}
		windows.WaitForSingleObject(h, 15000)
		windows.CloseHandle(h)
		break
	}
	if exe, err := executablePath(); err == nil {
		os.Remove(exe + ".old")
	}
}

func executablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// findAsset returns the exe of the release with the given tag. Only releases
// newer than the running version are accepted.
func findAsset(tag string) (ghAsset, error) {
	releases, err := fetchReleases()
	if err != nil {
		return ghAsset{}, err
	}
	curMajor, curMinor, curPatch, ok := parseSemver(CurrentVersion)
	if !ok {
		return ghAsset{}, errors.New("cannot parse current version")
	}
	for _, rel := range releases {
		if rel.Draft || rel.TagName != tag {
			continue
		}
		major, minor, patch, ok := parseSemver(rel.TagName)
		if !ok || !isNewer(major, minor, patch, curMajor, curMinor, curPatch) {
			break
		}
		for _, asset := range rel.Assets {
			if strings.HasSuffix(strings.ToLower(asset.Name), ".exe") &&
				strings.HasPrefix(asset.DownloadURL, downloadPrefix) {
				return asset, nil
			}
		}
		break
	}
	return ghAsset{}, fmt.Errorf("no update found for %s", tag)
}

// download saves the asset to path and checks it against the SHA-256 digest
// GitHub reports for it.
func download(asset ghAsset, path string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(asset.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(resp.Body, maxUpdateSize+1))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if n > maxUpdateSize {
		return errors.New("update is too large")
	}

	if want, ok := strings.CutPrefix(asset.Digest, "sha256:"); ok {
		if got := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(got, want) {
			return errors.New("update is damaged (checksum mismatch)")
		}
	}
	return nil
}
