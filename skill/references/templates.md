# HTML/Go Template Patterns

Reusable patterns for gohl + HTMLayout applications.

## Titlebar (Frameless Window)

```html
<div id="title-bar-wrap">
    <div id="title-bar" -gohl-drag>
        <div class="soft-name">App Name</div>
        <div>
            <button id="min-btn" class="window-btn" -gohl-min></button>
            <button id="close-btn" class="window-btn" -gohl-close></button>
        </div>
    </div>
</div>
```

```css
#title-bar-wrap { background: green green rgb(154,154,45) rgb(105,197,197); }
#title-bar {
    width: 100%; color: white;
    horizontal-align: right; vertical-align: middle;
    cursor: default; user-select: none;
    flow: horizontal;
}
.window-btn {
    width: 10dip; height: 10dip;
    border-radius: 50%; border: none;
    cursor: pointer; margin-right: 2dip;
    background: none;
}
#min-btn { background: #4b5454; }
#min-btn:hover { background: #4ecdc4; }
#close-btn { background: #a18582; }
#close-btn:hover { background: #e74c3c; }
```

## Modal Overlay (with role convention)

```html
<div class="modal-overlay" role="modal-overlay" id="default-modal">
    <div class="modal" role="modal-dialog">
        <div class="modal-header">
            <span role="modal-title">Title</span>
            <button class="modal-close" role="modal-close">×</button>
        </div>
        <div class="modal-body" role="modal-body">Content</div>
        <div class="modal-footer">
            <button class="btn btn-secondary" role="modal-cancel">Cancel</button>
            <button class="btn btn-primary" role="modal-ok">OK</button>
        </div>
    </div>
</div>
```

```css
.modal-overlay {
    position: fixed; top: 0; left: 0; right: 0; bottom: 0;
    display: none;
    horizontal-align: center; vertical-align: middle;
    background: rgba(0, 0, 0, 0.9);
    z-index: 2000;
}
.modal-overlay.show { display: block; }
.modal {
    behavior: light-box-dialog;
    background: #282B32; border-radius: 8dip;
    min-width: 300dip; max-width: 500dip;
}
```

## showModal Function (with callbacks)

```go
func showModal(id, title, body string,
    cbk func(submit bool) bool,
    onRender func(btnOk, btnCancel, btnClose *gohl.Element)) {

    overlay := gw.GetRootElement().GetElementById(id)
    titleEl := overlay.GetElementByAttr("role", "modal-title")
    bodyEl := overlay.GetElementByAttr("role", "modal-body")
    btnCancelEl := overlay.GetElementByAttr("role", "modal-cancel")
    btnOkEl := overlay.GetElementByAttr("role", "modal-ok")
    btnCloseEl := overlay.GetElementByAttr("role", "modal-close")

    if btnOkEl != nil { btnOkEl.Show() }
    if btnCancelEl != nil { btnCancelEl.Show() }
    if btnCloseEl != nil { btnCloseEl.Show() }

    if titleEl != nil { titleEl.SetText(title) }
    if bodyEl != nil && body != "" { bodyEl.SetHtml(body) }
    if overlay != nil { overlay.Show() }

    if onRender != nil && btnOkEl != nil && btnCancelEl != nil && btnCloseEl != nil {
        onRender(btnOkEl, btnCancelEl, btnCloseEl)
    }

    if btnCloseEl != nil {
        btnCloseEl.SetOnClick(func(owner *gohl.Element, params *gohl.BehaviorEventParams) bool {
            if cbk != nil { cbk(false) }
            overlay.Hide()
            return true
        })
    }
    if btnCancelEl != nil {
        btnCancelEl.SetOnClick(func(owner *gohl.Element, params *gohl.BehaviorEventParams) bool {
            if cbk != nil { cbk(false) }
            overlay.Hide()
            return true
        })
    }
    if btnOkEl != nil {
        btnOkEl.SetOnClick(func(owner *gohl.Element, params *gohl.BehaviorEventParams) bool {
            if cbk != nil && cbk(true) { overlay.Hide() }
            return true
        })
    }
}

// Usage
showModal("default-modal", "Confirm", "Delete this?", func(ok bool) bool {
    if ok { /* do delete */ }
    return true
}, nil)

// Custom button visibility
showModal("default-modal", "Info", "Done!", func(ok bool) bool {
    return true
}, func(btnOk, btnCancel, btnClose *gohl.Element) {
    btnOk.Show()
    btnCancel.Hide()
    btnClose.Hide()
})
```

## Toast Notification

```html
<div class="notification" id="notification">
    <span id="notification-text">Message</span>
</div>
```

```css
.notification {
    position: fixed; bottom: 20px;
    text-align: center;
    background: #aadb36; color: rgb(23, 24, 21);
    padding: 15px 20px; border-radius: 4px;
    display: none; z-index: 4000;
    width: 100%; margin: 0 10px;
}
.notification.show {
    display: block;
    assigned!: self::opacity = 0.01, self.start-animation();
    animation-step!: self::opacity < 1.0
        ? ( self::opacity = self::opacity + 0.01, return 10 )
        # ( self.fade = "", return cancel );
}
```

```go
func showNotification(text string, dismissTime time.Duration) {
    gw.UpdateUI(
        gohl.U{ID: "notification", Action: "show", Value: true},
        gohl.U{ID: "notification-text", Action: "text", Value: text},
    )
    time.AfterFunc(dismissTime, func() {
        gw.UpdateUI(gohl.U{ID: "notification", Action: "show", Value: false})
    })
}
```

## Sidebar + Main Layout

```html
<div id="app">
    <div id="sidebar">
        <div class="nav-item active" id="nav-library">Library</div>
        <div class="nav-item" id="nav-local">Local</div>
    </div>
    <div id="main"><!-- Main content --></div>
</div>
```

```css
#app { size: *; flow: horizontal; }
#sidebar {
    height: *; width: 200dip;
    flow: vertical;
    background-color: #0e1120;
    border-right: 1dip solid #2a3046;
}
#main { size: *; flow: vertical; }
```

## Page Switching

```go
func showPage(pageId string) {
    gw.Dispatch(func() {
        for _, page := range gw.GetRootElement().Select("#app>div") {
            if id, ok := page.Attr("id"); ok {
                if id == pageId {
                    page.AddClass("active")
                } else {
                    page.RemoveClass("active")
                }
            }
        }
    })
}
```

```css
.page { size: *; flow: vertical; display: none; }
.page.active { display: block; }
```

## Progress Bar

```html
<div class="progress">
    <div class="fill" id="dlfill" style="width:37%"></div>
</div>
```

```css
.progress { height: 5dip; border-radius: 3dip; background: #1b2034; overflow: hidden; }
.progress .fill { height: 100%; background-color: #e8a33c #ffd88a #ffd88a #e8a33c; }
```

## Helper Functions

```go
func setText(id, txt string) {
    if win == nil { return }
    if el := win.GetElementById(id); el != nil { el.SetText(txt) }
}

func setClass(id, cls string) {
    if win == nil { return }
    if el := win.GetElementById(id); el != nil { el.SetAttr("class", cls) }
}

func setAttr(id, attr, val string) {
    if win == nil { return }
    if el := win.GetElementById(id); el != nil { el.SetAttr(attr, val) }
}

func getInputValue(id string) string {
    if win == nil { return "" }
    if el := win.GetElementById(id); el != nil {
        v, _ := el.GetValue()
        return v
    }
    return ""
}

func show(id string, on bool) {
    if win == nil { return }
    if el := win.GetElementById(id); el != nil {
        if on { el.Show() } else { el.Hide() }
    }
}
```

## Embed Resource Loader

```go
//go:embed embed.zip
var embedZip embed.FS

var embedReader *zip.Reader

func initEmbedZip() {
    data, _ := embedZip.ReadFile("embed.zip")
    embedReader, _ = zip.NewReader(bytes.NewReader(data), int64(len(data)))
}

func loadEmbedRes(uri string) string {
    uri = strings.TrimPrefix(uri, "embed://")
    uri = strings.TrimSuffix(uri, "/")
    if embedReader == nil { return "" }
    for _, file := range embedReader.File {
        if file.Name == uri {
            rc, _ := file.Open()
            defer rc.Close()
            data, _ := io.ReadAll(rc)
            return string(data)
        }
    }
    return ""
}

func embedLoader(uri string) ([]byte, uint32, bool) {
    data := loadEmbedRes(uri)
    if data == "" { return nil, 0, false }
    filename := strings.TrimPrefix(uri, "embed://")
    return []byte(data), gohl.GetResourceDataType(filename), true
}

// Register in main():
gohl.RegisterResourceLoader("embed", embedLoader)
```

## Minimize to Tray

```go
const WM_TRAYMSG = 0x0400 + 1

tray = gohl.NewTrayIcon(gohl.TrayConfig{
    Icon: win.GetIcon(),
    Tip:  "App Name",
})

win.OnMinimize = func() bool {
    if !tray.IsAdded() {
        tray.Add(win.GetHwnd(), WM_TRAYMSG)
        tray.ShowInfo("Minimized", "Click tray icon to restore")
    }
    win.Hide()
    return false  // prevent default minimize
}
```

## Background Progress Update

```go
go func() {
    for pct := 0; pct <= 100; pct++ {
        time.Sleep(200 * time.Millisecond)
        win.Dispatch(func() {
            setText("progress-label", fmt.Sprintf("%d%%", pct))
            if el := win.GetElementById("progress-fill"); el != nil {
                el.SetStyle("width", fmt.Sprintf("%d%%", pct))
            }
        })
    }
}()
```
