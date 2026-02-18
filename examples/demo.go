package main

import (
	"gohl"
	"log"
	"syscall"
	"unsafe"
)

var gw *gohl.Window

// 托盘相关的常量和结构体
const (
	NIM_ADD        = 0x00000000
	NIM_MODIFY     = 0x00000001
	NIM_DELETE     = 0x00000002
	NIM_SETVERSION = 0x00000004
	NIF_MESSAGE    = 0x00000001
	NIF_ICON       = 0x00000002
	NIF_TIP        = 0x00000004
	NIF_INFO       = 0x00000010
	NIIF_INFO      = 0x00000001
	NIIF_USER      = 0x00000004
	WM_TRAYMSG     = 0x0400 + 1
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

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procShell_NotifyIcon = shell32.NewProc("Shell_NotifyIconW")
)

// 托盘图标是否已添加
var trayIconAdded bool

func Shell_NotifyIcon(dwMessage uint32, pnid *NOTIFYICONDATA) bool {
	ret, _, _ := procShell_NotifyIcon.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(pnid)))
	return ret != 0
}

// 添加托盘图标
func addTrayIcon() {
	if trayIconAdded {
		return
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = uintptr(gw.GetHwnd())
	nid.UId = 1
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UMsg = WM_TRAYMSG

	// 使用窗口配置中的图标
	hIcon := gw.GetIcon()
	log.Printf("[Tray] 窗口图标句柄: %v", hIcon)

	// 如果获取不到图标，使用默认值
	if hIcon == 0 {
		// 尝试从 shell32.dll 获取一个默认图标
		extractIcon := shell32.NewProc("ExtractIconW")
		if extractIcon != nil {
			hIcon, _, _ = extractIcon.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("shell32.dll"))), 4)
			log.Printf("[Tray] 从 shell32.dll 获取图标: %v", hIcon)
		}
	}

	nid.HIcon = hIcon

	// 设置提示文本
	tip := "Behaviors Demo"
	tipUtf16 := syscall.StringToUTF16(tip)
	copy(nid.SzTip[:], tipUtf16)

	if Shell_NotifyIcon(NIM_ADD, &nid) {
		trayIconAdded = true
		log.Println("[Tray] 托盘图标已添加")
	} else {
		log.Println("[Tray] 添加托盘图标失败")
	}
}

// 删除托盘图标
func removeTrayIcon() {
	if !trayIconAdded {
		return
	}

	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = uintptr(gw.GetHwnd())
	nid.UId = 1

	if Shell_NotifyIcon(NIM_DELETE, &nid) {
		trayIconAdded = false
		log.Println("[Tray] 托盘图标已删除")
	}
}

// 显示托盘提示
func showTrayTip(title, message string) {
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = uintptr(gw.GetHwnd())
	nid.UId = 1
	nid.UFlags = NIF_INFO

	// 设置提示消息
	msgUtf16 := syscall.StringToUTF16(message)
	copy(nid.SzInfo[:], msgUtf16)

	// 设置提示标题
	titleUtf16 := syscall.StringToUTF16(title)
	copy(nid.SzInfoTitle[:], titleUtf16)

	// 设置提示图标和超时
	nid.DwInfoFlags = NIIF_INFO
	nid.UTimeout = 3000 // 3秒

	if Shell_NotifyIcon(NIM_MODIFY, &nid) {
		log.Println("[Tray] 显示托盘提示成功:", title, "-", message)
	} else {
		log.Println("[Tray] 显示托盘提示失败")
	}
}

func main() {
	gw = gohl.NewWindow(gohl.WindowConfig{
		Title:        "Behaviors Demo",
		Width:        560,
		Height:       500,
		Frameless:    true,
		Resize:       false,
		Border:       true,
		Rounded:      true,
		CornerRadius: 10,
		Icon:         gohl.LoadIconFromResource(2),
		Center:       true,
	})

	gw.SetNotifyHandler(&gohl.NotifyHandler{
		Behaviors: map[string]*gohl.EventHandler{},
	})

	gw.OnClick = func(elem *gohl.Element) {
		id, _ := elem.Attr("id")
		switch id {
		case "show-modal":
			showModal("示例对话框", "这是一个模态对话框示例。\n点击确定或取消关闭对话框。")
		case "show-confirm":
			showModal("确认操作", "您确定要执行此操作吗？")
		case "show-alert":
			showModal("警告", "这是一个警告消息！\n请注意风险。")
		case "modal-close", "modal-cancel":
			hideModal()
			if id == "modal-cancel" {
				showNotification("已取消操作")
			}
		case "modal-ok":
			hideModal()
			showNotification("操作已确认")
		case "close-btn":
			// 关闭窗口前删除托盘图标
			removeTrayIcon()
		}
	}

	gw.OnMinimize = func() bool {
		// 添加托盘图标
		addTrayIcon()
		// 显示托盘提示
		showTrayTip("应用已最小化", "点击托盘图标恢复窗口")
		// 隐藏窗口（而不是最小化）
		gw.Hide()
		// 返回 false 阻止后续的最小化操作
		return false
	}

	gw.OnHyperlinkClick = func(elem *gohl.Element) {
		id, _ := elem.Attr("id")
		switch id {
		case "link1":
			showNotification("点击了: 链接 1 - 访问首页")
		case "link2":
			showNotification("点击了: 链接 2 - 查看文档")
		case "link3":
			showNotification("点击了: 链接 3 - 联系我们")
		default:
			showNotification("点击了超链接")
		}
	}

	gw.LoadFile("demo.html").Run()
}

func showModal(title, body string) {
	overlay := gw.GetRootElement().GetElementById("modal-overlay")
	titleEl := gw.GetRootElement().GetElementById("modal-title")
	bodyEl := gw.GetRootElement().GetElementById("modal-body")

	if titleEl != nil {
		titleEl.SetText(title)
	}
	if bodyEl != nil {
		bodyEl.SetHtml(body)
	}
	if overlay != nil {
		overlay.Show()
	}

	log.Println("[Modal] Shown:", title)
}

func hideModal() {
	overlay := gw.GetRootElement().GetElementById("modal-overlay")
	if overlay != nil {
		overlay.Hide()
	}
	log.Println("[Modal] Hidden")
}

func showNotification(text string) {
	notification := gw.GetRootElement().GetElementById("notification")
	textEl := gw.GetRootElement().GetElementById("notification-text")

	if notification != nil && textEl != nil {
		textEl.SetText(text)
		notification.Show()

		gw.SetTimer(2000, func() {
			log.Println("[Notification] Hidden after 2 seconds")
			hideNotification()
		})
	}

	log.Println("[Notification]", text)
}

func hideNotification() {
	notification := gw.GetRootElement().GetElementById("notification")
	if notification != nil {
		notification.Hide()
	}
}
