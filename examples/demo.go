package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/forbe/gohl"
)

//go:embed embed.zip
var embedZip embed.FS

var (
	embedReader *zip.Reader
	embedOnce   sync.Once
)

var gw *gohl.Window
var tray *gohl.TrayIcon

const WM_TRAYMSG = 0x0400 + 1

func main() {
	initEmbedZip()
	gohl.RegisterResourceLoader("embed", embedLoader)

	gw = gohl.NewWindow(gohl.WindowConfig{
		Title:        "Behaviors Demo",
		Width:        360,
		Height:       480,
		Frameless:    true,
		Resize:       false,
		Border:       true,
		Rounded:      true,
		CornerRadius: 10,
		Icon:         gohl.LoadIconFromResource(2),
		Center:       true,
	})

	store, _ := gohl.NewStorage("GameOnline")
	tray = gohl.NewTrayIcon(gohl.TrayConfig{
		Icon: gw.GetIcon(),
		Tip:  "Behaviors Demo",
	})

	gw.SetNotifyHandler(&gohl.NotifyHandler{
		Behaviors: map[string]*gohl.EventHandler{},
	})

	gw.OnClick = func(elem *gohl.Element) {
		role, _ := elem.Attr("role")
		id, _ := elem.Attr("id")
		switch role {
		case "show-modal":
			if id == "btn-create" {
				// showLoading(true)
				alert("请先创建游戏")
				return
			}
			if id == "btn-join" {
				html := `
					<div id="join-code-container">
						<input id="join-code" type="text" /><br />
					</div>
				`
				lastCode, ok := store.GetString("last-code")
				if ok {
					html = `
						<div id="join-code-container">
							<input id="join-code" type="text" /><br />
							✧ &nbsp;<a href="#" id="join-code-history">` + lastCode + `</a>
						</div>
					`
				}
				showModal("modal01", "请输入联机码", html)
				return
			}
			showModal("modal01", "示例对话框", "这是一个模态对话框示例。\n点击确定或取消关闭对话框。")
		case "show-confirm":
			showModal("modal01", "确认操作", "您确定要执行此操作吗？")
		case "show-alert":
			showModal("modal01", "警告", "这是一个警告消息！\n请注意风险。")
		case "modal-close", "modal-cancel":
			hideModal(elem)
			if role == "modal-cancel" {
				showNotification("已取消操作")
			}
		case "modal-ok":
			joinCode := gw.GetRootElement().GetElementById("join-code")
			if joinCode != nil {
				code := joinCode.Text()
				if len(code) < 6 {
					showNotification("不正确的联机码")
					return
				}
				store.Set("last-code", code)
			}
			hideModal(elem)
			showNotification("操作已确认")

		case "setting-menu":
			showModal("modal-setting", "设置", "")
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
		case "join-code-history":
			joinCode := gw.GetRootElement().GetElementById("join-code")
			if joinCode != nil {
				joinCode.SetText(elem.Text())
			}
		case "link2":
			showNotification("点击了: 链接 2 - 查看文档")
		case "link3":
			showNotification("点击了: 链接 3 - 联系我们")
		default:
			showNotification("点击了超链接")
		}
	}

	gw.LoadFile("demo2.html").Run()
}

func showLoading(show bool) {
	gw.UpdateUI(gohl.U{
		ID:     "loading",
		Action: "show",
		Value:  show,
	})
}

func alert(msg string) {
	showModal("modal01", "警告", msg)
}

func showModal(id, title, body string) {
	overlay := gw.GetRootElement().GetElementById(id)

	titleEl := overlay.GetElementByAttr("role", "modal-title")
	bodyEl := overlay.GetElementByAttr("role", "modal-body")

	if titleEl != nil {
		titleEl.SetText(title)
	}
	if bodyEl != nil {
		if body != "" {
			bodyEl.SetHtml(body)
		}
	}
	if overlay != nil {
		overlay.Show()
	}

	log.Println("[Modal] Shown:", title)
}

func hideModal(elem *gohl.Element) {
	overlay := elem.FindParentByAttr("role", "modal-overlay")
	if overlay != nil {
		overlay.Hide()
		return
	}
	log.Fatal("弹窗cover图层请设置 role='modal-overlay' 否则无法关闭!")
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

func initEmbedZip() {
	embedOnce.Do(func() {
		data, err := embedZip.ReadFile("embed.zip")
		if err != nil {
			log.Printf("Failed to read embed.zip: %v", err)
			return
		}
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			log.Printf("Failed to parse embed.zip: %v", err)
			return
		}
		embedReader = reader
	})
}

// 从embed.zip加载资源
func embedLoader(uri string) ([]byte, uint32, bool) {
	if embedReader == nil {
		return nil, 0, false
	}

	filename := uri[8:]
	filename = strings.TrimSuffix(filename, "/")

	for _, file := range embedReader.File {
		if file.Name == filename {
			rc, err := file.Open()
			if err != nil {
				return nil, 0, false
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, 0, false
			}

			return data, gohl.GetResourceDataType(filename), true
		}
	}

	return nil, 0, false
}
