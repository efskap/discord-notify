package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"path/filepath"
	"unsafe"
)

var (
	user32                       = windows.NewLazyDLL("user32.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	psapi                        = windows.NewLazyDLL("Psapi.dll")
	procGetProcessImageFileName  = psapi.NewProc("GetProcessImageFileNameW")
)

func isRipcordFocused() (bool, error) {
	// find handle of foreground window
	fgWinHandle, _, err := procGetForegroundWindow.Call()
	if fgWinHandle == 0 {
		return false, fmt.Errorf("could not get foreground window handle: %w", err)
	}

	// find process id of handle
	var dwProcessId uint32
	_, _, _ = procGetWindowThreadProcessId.Call(fgWinHandle, uintptr(unsafe.Pointer(&dwProcessId)))
	// open process by id
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, dwProcessId)
	if err != nil {
		return false, fmt.Errorf("error opening proc: %w", err)
	}
	defer windows.Close(handle)

	// find filepath of the process's exe
	buf := make([]uint16, 255)
	lenCopied, _, err := procGetProcessImageFileName.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)))

	if lenCopied == 0 {
		return false, fmt.Errorf("error getting proc filename: %w", err)
	}

	// we use the `W` variant of the function, so it's UTF16 encoded
	procPath := windows.UTF16ToString(buf)
	procName := filepath.Base(procPath)

	return procName == "Ripcord.exe", nil
}
