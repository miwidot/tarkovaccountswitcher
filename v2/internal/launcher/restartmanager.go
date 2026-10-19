package launcher

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	rstrtmgr                = windows.NewLazySystemDLL("rstrtmgr.dll")
	procRmStartSession      = rstrtmgr.NewProc("RmStartSession")
	procRmRegisterResources = rstrtmgr.NewProc("RmRegisterResources")
	procRmShutdown          = rstrtmgr.NewProc("RmShutdown")
	procRmEndSession        = rstrtmgr.NewProc("RmEndSession")
)

// rmUniqueProcess mirrors RM_UNIQUE_PROCESS.
type rmUniqueProcess struct {
	processID        uint32
	processStartTime windows.Filetime
}

// shutdownApps asks the given GUI processes to exit through the Restart
// Manager (WM_QUERYENDSESSION/WM_ENDSESSION with ENDSESSION_CLOSEAPP), without
// force-terminating them. Returns true if all of them shut down.
func shutdownApps(pids []uint32) bool {
	var procs []rmUniqueProcess
	for _, pid := range pids {
		h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			continue
		}
		var created, exited, kernel, user windows.Filetime
		err = windows.GetProcessTimes(h, &created, &exited, &kernel, &user)
		windows.CloseHandle(h)
		if err == nil {
			procs = append(procs, rmUniqueProcess{pid, created})
		}
	}
	if len(procs) == 0 {
		return false
	}

	var session uint32
	key := make([]uint16, 33) // CCH_RM_SESSION_KEY + 1
	if r, _, _ := procRmStartSession.Call(uintptr(unsafe.Pointer(&session)), 0, uintptr(unsafe.Pointer(&key[0]))); r != 0 {
		return false
	}
	defer procRmEndSession.Call(uintptr(session))

	if r, _, _ := procRmRegisterResources.Call(uintptr(session), 0, 0,
		uintptr(len(procs)), uintptr(unsafe.Pointer(&procs[0])), 0, 0); r != 0 {
		return false
	}
	r, _, _ := procRmShutdown.Call(uintptr(session), 0, 0)
	return r == 0
}
