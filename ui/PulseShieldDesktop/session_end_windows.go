//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WM_QUERYENDSESSION      = 0x0011
	WM_ENDSESSION           = 0x0016
	WM_DESTROY              = 0x0002
	WM_WTSSESSION_CHANGE    = 0x02B1
	WTS_SESSION_LOGOFF      = 0x6
	NOTIFY_FOR_THIS_SESSION = 0
	COLOR_WINDOW            = 5
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")
	wtsapi32 = windows.NewLazySystemDLL("wtsapi32.dll")

	procGetModuleHandleW                 = kernel32.NewProc("GetModuleHandleW")
	procRegisterClassExW                 = user32.NewProc("RegisterClassExW")
	procUnregisterClassW                 = user32.NewProc("UnregisterClassW")
	procCreateWindowExW                  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                   = user32.NewProc("DefWindowProcW")
	procGetMessageW                      = user32.NewProc("GetMessageW")
	procTranslateMessage                 = user32.NewProc("TranslateMessage")
	procDispatchMessageW                 = user32.NewProc("DispatchMessageW")
	procDestroyWindow                    = user32.NewProc("DestroyWindow")
	procWTSRegisterSessionNotification   = wtsapi32.NewProc("WTSRegisterSessionNotification")
	procWTSUnRegisterSessionNotification = wtsapi32.NewProc("WTSUnRegisterSessionNotification")
)

type wndClassExW struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   uintptr
	ClassName  uintptr
	IconSm     uintptr
}

type point struct {
	X, Y int32
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	Pad0    uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pad1    uint32
	Pt      point
}

var (
	sessionHookMu  sync.Mutex
	sessionHookApp *App
)

func startSessionEndNotifier(app *App) {
	sessionHookMu.Lock()
	sessionHookApp = app
	sessionHookMu.Unlock()

	go runSessionMessagePump()
}

func runSessionMessagePump() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cb := syscall.NewCallback(func(hwnd, msg, wParam, lParam uintptr) uintptr {
		return sessionWndProc(windows.HWND(hwnd), uint32(msg), wParam, lParam)
	})

	mod, _, _ := procGetModuleHandleW.Call(0)
	className := fmt.Sprintf("PulseShieldSessionHook_%d", os.Getpid())
	classPtr, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return
	}
	titlePtr, _ := windows.UTF16PtrFromString("")

	wc := wndClassExW{
		Size:       uint32(unsafe.Sizeof(wndClassExW{})),
		WndProc:    cb,
		Instance:   mod,
		Background: uintptr(COLOR_WINDOW + 1),
		ClassName:  uintptr(unsafe.Pointer(classPtr)),
	}
	r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return
	}
	defer func() {
		_, _, _ = procUnregisterClassW.Call(uintptr(unsafe.Pointer(classPtr)), mod)
	}()

	parent := ^uintptr(2) // HWND_MESSAGE (-3)
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(classPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0,
		0, 0, 0, 0,
		parent,
		0,
		mod,
		0,
	)
	if hwnd == 0 {
		return
	}
	defer func() {
		_, _, _ = procWTSUnRegisterSessionNotification.Call(hwnd)
		_, _, _ = procDestroyWindow.Call(hwnd)
	}()

	_, _, _ = procWTSRegisterSessionNotification.Call(hwnd, NOTIFY_FOR_THIS_SESSION)

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == 0 {
			break
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func sessionWndProc(hwnd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_QUERYENDSESSION:
		runShutdownCleanupFromSession()
		return 1
	case WM_ENDSESSION:
		if wParam != 0 {
			runShutdownCleanupFromSession()
		}
		return 0
	case WM_WTSSESSION_CHANGE:
		if uint32(wParam) == WTS_SESSION_LOGOFF {
			runShutdownCleanupFromSession()
		}
		return 0
	case WM_DESTROY:
		return 0
	default:
		r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
		return r
	}
}

func runShutdownCleanupFromSession() {
	sessionHookMu.Lock()
	app := sessionHookApp
	sessionHookMu.Unlock()
	if app != nil {
		app.ShutdownCleanup()
	}
}
