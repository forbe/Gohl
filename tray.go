package gohl

import (
	"syscall"
	"unsafe"
)

const (
	NIM_ADD              = 0x00000000
	NIM_MODIFY           = 0x00000001
	NIM_DELETE           = 0x00000002
	NIM_SETVERSION       = 0x00000004
	NIF_MESSAGE          = 0x00000001
	NIF_ICON             = 0x00000002
	NIF_TIP              = 0x00000004
	NIF_INFO             = 0x00000010
	NIF_SHOWTIP          = 0x00000080
	NIIF_NONE            = 0x00000000
	NIIF_INFO            = 0x00000001
	NIIF_WARNING         = 0x00000002
	NIIF_ERROR           = 0x00000003
	NIIF_USER            = 0x00000004
	NIIF_NOSOUND         = 0x00000010
	NIIF_LARGE_ICON      = 0x00000020
	NOTIFYICON_VERSION   = 0x00000003
	NOTIFYICON_VERSION_4 = 0x00000004
)

type NOTIFYICONDATA struct {
	CbSize        uint32
	Hwnd          uintptr
	UId           uint32
	UFlags        uint32
	UMsg          uint32
	HIcon         uintptr
	SzTip         [128]uint16
	DwState       uint32
	DwStateMask   uint32
	SzInfo        [256]uint16
	UTimeout      uint32
	UVersion      uint32
	SzInfoTitle   [64]uint16
	DwInfoFlags   uint32
	SzBalloonIcon [256]uint16
}

type TrayIcon struct {
	hwnd  uintptr
	uId   uint32
	hIcon uintptr
	tip   string
	added bool
}

type TrayConfig struct {
	Icon    uintptr
	Tip     string
	UId     uint32
	OnClick func()
}

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procShell_NotifyIcon = shell32.NewProc("Shell_NotifyIconW")
	procExtractIcon      = shell32.NewProc("ExtractIconW")
)

func shellNotifyIcon(dwMessage uint32, pnid *NOTIFYICONDATA) bool {
	ret, _, _ := procShell_NotifyIcon.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(pnid)))
	return ret != 0
}

func extractIconFromFile(filePath string, index int) uintptr {
	ret, _, _ := procExtractIcon.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filePath))), uintptr(index))
	return ret
}

func NewTrayIcon(config TrayConfig) *TrayIcon {
	tray := &TrayIcon{
		hIcon: config.Icon,
		tip:   config.Tip,
		uId:   config.UId,
	}
	if tray.uId == 0 {
		tray.uId = 1
	}
	return tray
}

func (t *TrayIcon) Add(hwnd uintptr, msg uint32) bool {
	if t.added {
		return true
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = hwnd
	nid.UId = t.uId
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UMsg = msg
	nid.HIcon = t.hIcon

	if t.tip != "" {
		tipUtf16 := syscall.StringToUTF16(t.tip)
		copy(nid.SzTip[:], tipUtf16)
	}

	if !shellNotifyIcon(NIM_ADD, &nid) {
		return false
	}

	nid.UVersion = NOTIFYICON_VERSION_4
	shellNotifyIcon(NIM_SETVERSION, &nid)

	t.hwnd = hwnd
	t.added = true
	return true
}

func (t *TrayIcon) Remove() bool {
	if !t.added {
		return true
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = t.hwnd
	nid.UId = t.uId

	if shellNotifyIcon(NIM_DELETE, &nid) {
		t.added = false
		return true
	}
	return false
}

func (t *TrayIcon) SetIcon(icon uintptr) bool {
	if !t.added {
		return false
	}

	t.hIcon = icon

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = t.hwnd
	nid.UId = t.uId
	nid.UFlags = NIF_ICON
	nid.HIcon = icon

	return shellNotifyIcon(NIM_MODIFY, &nid)
}

func (t *TrayIcon) SetTip(tip string) bool {
	if !t.added {
		return false
	}

	t.tip = tip

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = t.hwnd
	nid.UId = t.uId
	nid.UFlags = NIF_TIP

	tipUtf16 := syscall.StringToUTF16(tip)
	copy(nid.SzTip[:], tipUtf16)

	return shellNotifyIcon(NIM_MODIFY, &nid)
}

func (t *TrayIcon) ShowBalloon(title, message string, flags uint32, timeout uint32) bool {
	if !t.added {
		return false
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = t.hwnd
	nid.UId = t.uId
	nid.UFlags = NIF_INFO | NIF_SHOWTIP

	msgUtf16 := syscall.StringToUTF16(message)
	copy(nid.SzInfo[:], msgUtf16)

	titleUtf16 := syscall.StringToUTF16(title)
	copy(nid.SzInfoTitle[:], titleUtf16)

	nid.DwInfoFlags = flags
	nid.UTimeout = timeout

	return shellNotifyIcon(NIM_MODIFY, &nid)
}

func (t *TrayIcon) ShowInfo(title, message string) bool {
	return t.ShowBalloon(title, message, NIIF_INFO, 3000)
}

func (t *TrayIcon) ShowWarning(title, message string) bool {
	return t.ShowBalloon(title, message, NIIF_WARNING, 3000)
}

func (t *TrayIcon) ShowError(title, message string) bool {
	return t.ShowBalloon(title, message, NIIF_ERROR, 3000)
}

func (t *TrayIcon) IsAdded() bool {
	return t.added
}

func ExtractIcon(filePath string, index int) uintptr {
	return extractIconFromFile(filePath, index)
}

func GetDefaultIcon() uintptr {
	icon := extractIconFromFile("shell32.dll", 4)
	if icon == 0 {
		icon = 1
	}
	return icon
}
