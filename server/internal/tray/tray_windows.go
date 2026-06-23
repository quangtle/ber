//go:build windows

package tray

import (
	"encoding/binary"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

var (
	user32  = syscall.NewLazyDLL("user32.dll")
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procLoadIconW              = user32.NewProc("LoadIconW")
	procPostMessageW           = user32.NewProc("PostMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procCreateIconFromResource = user32.NewProc("CreateIconFromResource")

	procShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
)

const (
	wmCreate         = 0x0001
	wmDestroy        = 0x0002
	wmCommand        = 0x0111
	wmNull           = 0x0000
	wmApp            = 0x8000
	wmTrayCallback   = wmApp + 1

	// ponytail: latest NOTIFYICONDATA version on Win10+
	nidSize    = 1032
	nimAdd     = 0
	nimDelete  = 1
	nimModify  = 2
	nifMessage = 0x0001
	nifIcon    = 0x0002
	nifTip     = 0x0004

	mfSeparator = 0x0800
	mfString    = 0x0000

	tpmLeftAlign   = 0x0000
	tpmRightButton = 0x0002
)

// wndProc is the message handler for our hidden window.
// It must be a package-level variable to keep the callback from being GC'd.
var wndProc uintptr

// Tray is defined in tray.go; Run is the platform-specific entry point.

// Run creates the tray icon and enters the Windows message loop.
// It blocks until Quit is called or the window is destroyed.
func Run(icon []byte, tooltip string, onReady func(*Tray)) {
	t := &Tray{icon: icon, tooltip: tooltip, quit: make(chan struct{}, 1)}

	wndProc = syscall.NewCallback(func(hwnd, msg, wparam, lparam uintptr) uintptr {
		return trayWndProc(t, hwnd, msg, wparam, lparam)
	})

	className, _ := syscall.UTF16PtrFromString("ber_tray_window")

	hInst, _ := kernel32GetModuleHandle()

	wcex := make([]byte, unsafe.Sizeof(WNDCLASSEX{}))
	setWndClassEx(wcex, className, hInst)

	classAtom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcex[0])))
	if classAtom == 0 {
		// ponytail: class already registered from a previous instance — ignore
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0, // dwExStyle
		uintptr(unsafe.Pointer(className)), // lpClassName
		0, // lpWindowName — no title needed for hidden window
		0, // dwStyle
		0, 0, 0, 0, // x, y, w, h
		0,    // hWndParent
		0,    // hMenu
		hInst,
		0, // lpParam
	)
	if hwnd == 0 {
		// ponytail: window creation failed (e.g. non-interactive session),
		// fall back to signal-based blocking
		onReady(t)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		return
	}

	onReady(t)

	// Enter message loop
	var msg [48]byte // MSG struct on x64
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg[0])),
			0, 0, 0,
		)
		if ret == 0 { // WM_QUIT
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg[0])))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg[0])))
	}

	// Cleanup
	nid := newNotifyIconData()
	setNidHwnd(nid, hwnd)
	nimDeleteData(nid)
	procDestroyWindow.Call(hwnd)
}

func trayWndProc(t *Tray, hwnd, msg, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmCreate:
		t.addIcon(hwnd)
		return 0

	case wmDestroy:
		t.clearNotifyIcon(hwnd)
		procPostQuitMessage.Call(0)
		return 0

	case wmCommand:
		id := uint32(wparam)
		for _, item := range t.items {
			if item.ID == id && !item.disabled {
				select {
				case item.clicked <- struct{}{}:
				default:
				}
				break
			}
		}
		return 0

	case wmTrayCallback:
		switch lparam {
		case 0x007B, 0x0205: // WM_CONTEXTMENU / WM_RBUTTONUP (right-click)
			t.showMenu(hwnd)
		case 0x0203: // WM_LBUTTONDBLCLK
			if len(t.items) > 0 {
				select {
				case t.items[0].clicked <- struct{}{}:
				default:
				}
			}
		}
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wparam, lparam)
	return ret
}

func (t *Tray) addIcon(hwnd uintptr) {
	nid := newNotifyIconData()
	setNidHwnd(nid, hwnd)

	// ponytail: generate an HICON from embedded .ico, fallback to default
	hicon := icoToHICON(t.icon)
	if hicon == 0 {
		hicon, _, _ = procLoadIconW.Call(0, 32512) // IDI_APPLICATION
	}
	setNidIcon(nid, hicon)

	tip := t.tooltip
	if len(tip) > 127 {
		tip = tip[:127]
	}
	setNidTip(nid, tip)
	setNidFlags(nid, nifMessage|nifIcon|nifTip)
	setNidCallback(nid, wmTrayCallback)

	procShellNotifyIcon.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(nid)))
}

func (t *Tray) clearNotifyIcon(hwnd uintptr) {
	nid := newNotifyIconData()
	setNidHwnd(nid, hwnd)
	nimDeleteData(nid)
}

func (t *Tray) showMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}

	for _, item := range t.items {
		if item.Title == "-" {
			procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)
		} else {
			title, _ := syscall.UTF16PtrFromString(item.Title)
			procAppendMenuW.Call(hMenu, mfString, uintptr(item.ID), uintptr(unsafe.Pointer(title)))
		}
	}

	procSetForegroundWindow.Call(hwnd)

	var pt [2]int32
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt[0])))

	procTrackPopupMenu.Call(
		hMenu,
		tpmLeftAlign|tpmRightButton,
		uintptr(pt[0]), uintptr(pt[1]), 0, hwnd, 0,
	)
	procPostMessageW.Call(hwnd, wmNull, 0, 0)
	procDestroyMenu.Call(hMenu)
}

func (t *Tray) Quit() {
	procPostQuitMessage.Call(0)
}

func icoToHICON(ico []byte) uintptr {
	if len(ico) < 6 {
		return 0
	}
	count := int(ico[4]) | int(ico[5])<<8
	if count == 0 || len(ico) < 6+count*16 {
		return 0
	}
	entry := ico[6 : 6+16]
	sz := uint32(entry[8]) | uint32(entry[9])<<8 | uint32(entry[10])<<16 | uint32(entry[11])<<24
	off := uint32(entry[12]) | uint32(entry[13])<<8 | uint32(entry[14])<<16 | uint32(entry[15])<<24
	if int(off+sz) > len(ico) {
		return 0
	}
	data := ico[off : off+sz]
	h, _, _ := procCreateIconFromResource.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(sz),
		1, // fIcon = TRUE
		0x00030000, // dwVersion
	)
	return h
}

// --- WNDCLASSEX helpers ---------------------------------------------------

type WNDCLASSEX struct {
	_ [80]byte // approximate size for Win10 x64
}

func setWndClassEx(buf []byte, className *uint16, hInst uintptr) {
	// WNDCLASSEXW field offsets on x64:
	//   [0:4]   cbSize
	//   [4:8]   style
	//   [8:16]  lpfnWndProc
	//   [16:20] cbClsExtra
	//   [20:24] cbWndExtra
	//   [24:32] hInstance
	//   [32:40] hIcon
	//   [40:48] hCursor
	//   [48:56] hbrBackground
	//   [56:64] lpszMenuName
	//   [64:72] lpszClassName
	//   [72:80] hIconSm
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(buf)))
	*(*uintptr)(unsafe.Pointer(&buf[8])) = wndProc        // lpfnWndProc
	*(*uintptr)(unsafe.Pointer(&buf[24])) = hInst          // hInstance
	*(*uintptr)(unsafe.Pointer(&buf[64])) = uintptr(unsafe.Pointer(className)) // lpszClassName
}

// --- NOTIFYICONDATA helpers -----------------------------------------------

// notifyIconData is a fixed-size buffer that mirrors NOTIFYICONDATAW.
// Field offsets on x64:
//
//	[0:4]   cbSize
//	[8:16]  hWnd
//	[16:20] uID
//	[20:24] uFlags
//	[24:28] uCallbackMessage
//	[32:40] hIcon
//	[40:296] szTip[128]
type notifyIconData [nidSize]byte

func newNotifyIconData() *notifyIconData {
	n := new(notifyIconData)
	binary.LittleEndian.PutUint32(n[0:4], nidSize)
	binary.LittleEndian.PutUint32(n[16:20], 0) // uID = 0
	return n
}

func setNidHwnd(n *notifyIconData, hwnd uintptr) {
	binary.LittleEndian.PutUint64(n[8:16], uint64(hwnd))
}

func setNidFlags(n *notifyIconData, flags uint32) {
	binary.LittleEndian.PutUint32(n[20:24], flags)
}

func setNidCallback(n *notifyIconData, msg uint32) {
	binary.LittleEndian.PutUint32(n[24:28], msg)
}

func setNidIcon(n *notifyIconData, hicon uintptr) {
	binary.LittleEndian.PutUint64(n[32:40], uint64(hicon))
}

func setNidTip(n *notifyIconData, tip string) {
	buf := n[40:296]
	for i := 0; i < len(tip) && i < 127; i++ {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(tip[i]))
	}
}

func nimDeleteData(n *notifyIconData) {
	procShellNotifyIcon.Call(uintptr(nimDelete), uintptr(unsafe.Pointer(n)))
}

// kernel32 helpers

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

func kernel32GetModuleHandle() (uintptr, error) {
	ret, _, err := procGetModuleHandleW.Call(0)
	return ret, err
}

var procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
