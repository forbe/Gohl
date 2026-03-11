package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/forbe/gohl"
)

//go:embed embed.zip
var embedZip embed.FS

var (
	embedReader *zip.Reader
	embedOnce   sync.Once
)

var W *gohl.Window

const WM_TRAYMSG = 0x0400 + 1

func main() {
	initEmbedZip()
	gohl.RegisterResourceLoader("embed", embedLoader)

	W = gohl.NewWindow(gohl.WindowConfig{
		Title:        "联机助手",
		Width:        360,
		Height:       480,
		Frameless:    true,
		Resize:       false,
		Border:       true,
		Rounded:      true,
		CornerRadius: 10,
		Icon:         gohl.LoadIconFromResource(2),
		Center:       true,
		Handler: gohl.NotifyHandler{
			OnDocumentComplete: func() uintptr {
				log.Printf("OnDocumentComplete: %v", W.GetHwnd())
				//弹窗
				demoBtn := W.GetElementById("demo-btn")
				if demoBtn != nil {
					demoBtn.OnClick = func(elem *gohl.Element) bool {
						showModal("default-modal", "提示", "你好，欢迎来到中国!", func(submit bool) bool {
							return true
						}, nil)
						return true
					}
				}

				//弹窗
				demoBtn2 := W.GetElementById("demo-btn2")
				if demoBtn2 != nil {
					demoBtn2.OnClick = func(elem *gohl.Element) bool {
						W.UpdateUI(gohl.U{
							ID:     "demo-img",
							Action: "attr",
							Value:  map[string]interface{}{"src": "http://bd.kuaishoua.com/787b568727c8f5ce.png"},
						}, gohl.U{
							ID:     "demo-tips",
							Action: "text",
							Value:  "你好，欢迎来到中国!",
						})

						showNotification("已更新UI", 2*time.Second)
						return true
					}
				}

				return W.GetHwnd()
			},
		},
	})

	//窗体最小化之前调用，返回false不最小化
	W.OnMinimize = func() bool {
		log.Println("最小化窗口")
		return true
	}

	//点击按钮触发
	W.OnButtonClick = func(elem *gohl.Element) bool {
		role, _ := elem.Attr("role")
		id, _ := elem.Attr("id")
		switch role {
		case "show-modal":
			if id == "show-modal" {
				html := `<div id="join-code-container"><input id="join-code" type="text" /><br /></div>`
				showModal("default-modal", "请输入联机码", html, func(submit bool) bool {
					return true
				}, nil)
				return true
			}
		}
		return true
	}

	W.OnButtonStateChanged = func(elem *gohl.Element, checked bool) bool {
		log.Printf("OnButtonStateChanged: %v, %v", elem, checked)
		return true
	}

	W.OnHyperlinkClick = func(elem *gohl.Element) bool {
		href, ok := elem.Attr("@href")
		if ok {
			log.Println("点击了超链接: " + href)
			exec.Command("cmd", "/c", "start", href).Start()
		}
		return true
	}

	if FileExists("app.html") {
		W.LoadFile("app.html").Run()
	} else {
		html := loadEmbedRes("embed://app.html")
		W.SetHtml(html).Run()
	}
}

func showLoading(show bool) {
	W.UpdateUI(gohl.U{ID: "loading", Action: "show", Value: show})
}

func showModal(id, title, body string, cbk func(submit bool) bool, onRender func(btnOk, btnCancel, btnClose *gohl.Element)) {
	root := W.GetRootElement()
	if root == nil {
		return
	}
	overlay := root.GetElementById(id)
	if overlay == nil {
		return
	}
	titleEl := overlay.GetElementByAttr("role", "modal-title")
	bodyEl := overlay.GetElementByAttr("role", "modal-body")
	btnCancelEl := overlay.GetElementByAttr("role", "modal-cancel")
	btnOkEl := overlay.GetElementByAttr("role", "modal-ok")
	btnCloseEl := overlay.GetElementByAttr("role", "modal-close")
	if btnOkEl != nil {
		btnOkEl.Show()
	}
	if btnCancelEl != nil {
		btnCancelEl.Show()
	}
	if btnCloseEl != nil {
		btnCloseEl.Show()
	}

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

	if onRender != nil {
		if btnOkEl != nil && btnCancelEl != nil && btnCloseEl != nil {
			onRender(btnOkEl, btnCancelEl, btnCloseEl)
		}
	}

	if btnCloseEl != nil {
		btnCloseEl.OnClick = func(elem *gohl.Element) bool {
			if cbk != nil {
				cbk(false)
			}
			overlay.Hide()
			return true
		}
	}
	if btnCancelEl != nil {
		btnCancelEl.OnClick = func(elem *gohl.Element) bool {
			if cbk != nil {
				cbk(false)
			}
			overlay.Hide()
			return true
		}
	}
	if btnOkEl != nil {
		btnOkEl.OnClick = func(elem *gohl.Element) bool {
			if cbk != nil {
				if cbk(true) {
					overlay.Hide()
				}
			}
			return true
		}
	}
}

func showNotification(text string, dismissTime time.Duration) {
	W.UpdateUI(
		gohl.U{
			ID:     "notification",
			Action: "show",
			Value:  true,
		},
		gohl.U{
			ID:     "notification-text",
			Action: "text",
			Value:  text,
		},
	)
	time.AfterFunc(dismissTime, func() {
		hideNotification()
	})
}

func hideNotification() {
	W.UpdateUI(gohl.U{
		ID:     "notification",
		Action: "show",
		Value:  false,
	})
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

func loadEmbedRes(uri string) string {
	uri = strings.TrimPrefix(uri, "embed://")
	uri = strings.TrimSuffix(uri, "/")

	if embedReader == nil {
		return ""
	}

	for _, file := range embedReader.File {
		if file.Name == uri {
			rc, err := file.Open()
			if err != nil {
				return ""
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}
	return ""
}

func embedLoader(uri string) ([]byte, uint32, bool) {
	data := loadEmbedRes(uri)
	if data == "" {
		return nil, 0, false
	}
	filename := strings.TrimPrefix(uri, "embed://")
	return []byte(data), gohl.GetResourceDataType(filename), true
}

func alert(msg string) {
	showModal("default-modal", "提示", msg, func(submmit bool) bool {
		return true
	}, func(btnOk, btnCancel, btnClose *gohl.Element) {
		btnOk.Show()
		btnCancel.Hide()
		btnClose.Show()
	})
}

func showNewWindow(uri string, width, height int) {
	w := gohl.NewWindow(gohl.WindowConfig{
		Title:        "联机助手",
		Width:        width,
		Height:       height,
		Resize:       false,
		Border:       true,
		Frameless:    false,
		Rounded:      true,
		CornerRadius: 10,
		ClassName:    "HTMLayoutWindow" + strconv.Itoa(rand.Intn(1000000)),
		Icon:         gohl.LoadIconFromResource(2),
		Center:       true,
	})
	w.SetHtml(loadEmbedRes(uri)).Run()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !info.IsDir()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
