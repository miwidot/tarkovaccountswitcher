package updater

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// signerName is the certificate subject every release is signed with. An
// update signed by anyone else is rejected, even if its signature is valid.
const signerName = "Open Source Developer Martin Wilke"

var (
	wintrust                           = windows.NewLazySystemDLL("wintrust.dll")
	procWTHelperProvDataFromStateData  = wintrust.NewProc("WTHelperProvDataFromStateData")
	procWTHelperGetProvSignerFromChain = wintrust.NewProc("WTHelperGetProvSignerFromChain")
	procWTHelperGetProvCertFromChain   = wintrust.NewProc("WTHelperGetProvCertFromChain")
)

// cryptProviderCert mirrors the start of CRYPT_PROVIDER_CERT.
type cryptProviderCert struct {
	size uint32
	cert *windows.CertContext
}

// verifySignature checks the Authenticode signature of the file the way
// Windows does and returns the name of its signer.
func verifySignature(path string) (string, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	fileInfo := windows.WinTrustFileInfo{
		Size:     uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})),
		FilePath: pathPtr,
	}
	data := windows.WinTrustData{
		Size:                            uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice:                        windows.WTD_UI_NONE,
		RevocationChecks:                windows.WTD_REVOKE_NONE,
		UnionChoice:                     windows.WTD_CHOICE_FILE,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&fileInfo),
		StateAction:                     windows.WTD_STATEACTION_VERIFY,
	}
	action := windows.WINTRUST_ACTION_GENERIC_VERIFY_V2
	verifyErr := windows.WinVerifyTrustEx(windows.InvalidHWND, &action, &data)
	defer func() {
		data.StateAction = windows.WTD_STATEACTION_CLOSE
		windows.WinVerifyTrustEx(windows.InvalidHWND, &action, &data)
	}()
	if verifyErr != nil {
		return "", fmt.Errorf("update has no valid signature: %w", verifyErr)
	}

	// The helpers return pointers into the verification state, which stays
	// valid until the deferred close above.
	provData, _, _ := procWTHelperProvDataFromStateData.Call(uintptr(data.StateData))
	if provData == 0 {
		return "", errors.New("cannot read update signature")
	}
	signer, _, _ := procWTHelperGetProvSignerFromChain.Call(provData, 0, 0, 0)
	if signer == 0 {
		return "", errors.New("cannot read update signature")
	}
	provCert, _, _ := procWTHelperGetProvCertFromChain.Call(signer, 0)
	if provCert == 0 {
		return "", errors.New("cannot read update signature")
	}
	cert := (*(**cryptProviderCert)(unsafe.Pointer(&provCert))).cert
	if cert == nil {
		return "", errors.New("cannot read update signature")
	}

	name := make([]uint16, 256)
	windows.CertGetNameString(cert, windows.CERT_NAME_SIMPLE_DISPLAY_TYPE, 0, nil, &name[0], uint32(len(name)))
	return windows.UTF16ToString(name), nil
}
