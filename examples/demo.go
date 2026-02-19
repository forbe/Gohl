package main

import (
	"log"

	"github.com/forbe/gohl"
)

var gw *gohl.Window
var tray *gohl.TrayIcon

const WM_TRAYMSG = 0x0400 + 1

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

	tray = gohl.NewTrayIcon(gohl.TrayConfig{
		Icon: gw.GetIcon(),
		Tip:  "Behaviors Demo",
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
			if tray.IsAdded() {
				tray.Remove()
			}
		}
	}

	gw.OnMinimize = func() bool {
		if !tray.IsAdded() {
			tray.Add(gw.GetHwnd(), WM_TRAYMSG)
			tray.ShowInfo("应用已最小化", "点击托盘图标恢复窗口")
		}
		gw.Hide()
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
