# gohl API Complete Reference

> Source: `D:\codes\go\Gohl`

## Window

### WindowConfig

```go
type WindowConfig struct {
    Title        string        // Window title
    Width        int           // Width (logical px, auto-scaled by DPI)
    Height       int           // Height
    ClassName    string        // Window class name, default "HTMLayoutWindow"
    Border       bool          // Custom border
    Frameless    bool          // Frameless mode
    MaxBtn       bool          // Show maximize button
    MinBtn       bool          // Show minimize button
    Resize       bool          // Resizable
    Center       bool          // Center on screen
    Icon         uintptr       // HICON, use gohl.LoadIconFromResource(id)
    Rounded      bool          // Rounded corners
    CornerRadius int           // Corner radius, default 10
    Handler      NotifyHandler // Notification handler
}
```

### Window Creation & Control

```go
win := gohl.NewWindow(config)

// Load content
win.LoadFile("index.html")     // External file (auto-converts to file:/// absolute path)
win.SetHtml(htmlString)         // Set HTML string directly

// Chain calls
win.SetHtml(html).Run()
win.LoadFile("index.html").Run()

// Window control
win.Run()              // Enter message loop (blocking, call last)
win.Minimize()
win.Maximize()
win.Restore()
win.Close()            // posts WM_CLOSE (clean teardown, see SKILL.md rule 20)
win.Show()
win.Hide()
win.SetTitle("title")

// Window info
win.GetHwnd()          // uintptr - window handle
win.GetIcon()          // uintptr - icon handle
win.GetRootElement()   // *Element - root DOM element
win.GetElementById(id) // *Element - find by ID
```

### Window Event Callbacks

```go
// Button click (most common)
win.OnButtonClick = func(elem *gohl.Element, params *gohl.BehaviorEventParams) bool {
    role, _ := elem.Attr("role")
    id, _ := elem.Attr("id")
    return false // true=handled, false=propagate
}

// Checkbox/radio state change
win.OnButtonStateChanged = func(elem *gohl.Element, checked bool) bool { return false }

// Hyperlink click
win.OnHyperlinkClick = func(elem *gohl.Element) bool { return true }

// Select change
win.OnSelectionChanged = func(elem *gohl.Element, value string) bool { return false }

// Input value change
win.OnValueChange = func(elem *gohl.Element, value string) bool { return false }

// Visibility change
win.OnVisibleChange = func(elem *gohl.Element, visible bool) bool { return false }

// Mouse events
win.OnMouse = func(elem *gohl.Element, params *gohl.MouseParams) bool { return false }

// Minimize intercept (return false to prevent default minimize)
win.OnMinimize = func() bool { return false }
```

### NotifyHandler

```go
type NotifyHandler struct {
    Behaviors          map[string]*EventHandler  // Custom behavior mapping
    OnCreateControl    func(params *NmhlCreateControl) uintptr
    OnControlCreated   func(params *NmhlCreateControl) uintptr
    OnDestroyControl   func(params *NmhlDestroyControl) uintptr
    OnLoadData         func(params *NmhlLoadData) uintptr      // Resource loading intercept
    OnDataLoaded       func(params *NmhlDataLoaded) uintptr
    OnDocumentComplete func() uintptr                           // DOM loaded callback
}
```

### EventHandler

```go
type EventHandler struct {
    OnAttached      func(he HELEMENT)
    OnDetached      func(he HELEMENT)
    OnMouse         func(he HELEMENT, params *MouseParams) bool
    OnKey           func(he HELEMENT, params *KeyParams) bool
    OnFocus         func(he HELEMENT, params *FocusParams) bool
    OnDraw          func(he HELEMENT, params *DrawParams) bool
    OnTimer         func(he HELEMENT, params *TimerParams) bool
    OnBehaviorEvent func(he HELEMENT, params *BehaviorEventParams) bool
    OnMethodCall    func(he HELEMENT, params *MethodParams) bool
    OnDataArrived   func(he HELEMENT, params *DataArrivedParams) bool
    OnSize          func(he HELEMENT)
    OnScroll        func(he HELEMENT, params *ScrollParams) bool
    OnExchange      func(he HELEMENT, params *ExchangeParams) bool
    OnGesture       func(he HELEMENT, params *GestureParams) bool
}
```

## Element

### Find Elements

```go
win.GetElementById("myId")
win.GetRootElement()

el.GetElementById("id")
el.GetElementByAttr("role", "modal-title")
el.SelectFirst(".class")             // CSS selector, first match
el.SelectUnique("div#id")            // CSS selector, must be unique
el.Select("[panel]")                 // CSS selector, all matches
```

### Attributes

```go
el.Attr("id")                        // (string, bool)
el.SetAttr("class", "active")
el.RemoveAttr("disabled")
```

### Text & HTML

```go
el.Text()                            // string - text content
el.SetText("text")                   // WARNING: empty string panics!
el.Html()                            // string - innerHTML
el.SetHtml("<b>bold</b>")            // Replace innerHTML (must be compact single-line!)
el.OuterHtml()                       // string - outerHTML
el.PrependHtml("<div>before</div>")  // Insert at beginning
el.AppendHtml("<div>after</div>")    // Insert at end
```

### Form Values

```go
el.GetValue()                        // (string, int) - for input/select/textarea
el.SetValue("value")
el.ValueAsString()                   // (string, error) - alternative
```

> **Important:** `<input>` values must use `GetValue()`, `ValueAsString()` may return empty.

### Styles

```go
el.Style("display")                  // (string, bool)
el.SetStyle("display", "block")
el.RemoveStyle("display")
```

### Classes

```go
el.AddClass("active")
el.RemoveClass("active")
el.HasClass("active")                // bool
```

### Show/Hide

```go
el.Show()   // Adds "show" class + removes display:none + Update
el.Hide()   // Removes "show" class + sets display:none
```

> Show/Hide manipulates "show" class. Use CSS `.element.show { display: block; }` to配合.

### State

```go
el.SetState(STATE_CURRENT, true)
el.SetState(STATE_EXPANDED, true)
el.SetState(STATE_COLLAPSED, false)
el.SetState(STATE_FOCUS, true)
el.SetState(STATE_DISABLED, true)
el.State(STATE_CURRENT)              // bool
el.StateFlags()                      // uint32
el.SetStateFlags(flags)
el.IsChecked()                       // bool - checkbox/radio checked
el.IsVisible()                       // bool
```

### DOM Tree Traversal

```go
el.Parent()                          // *Element
el.Child(index)                      // *Element (0-based)
el.Children()                        // []*Element
el.ChildCount()                      // int
el.Index()                           // uint - index in parent
el.Root()                            // *Element
el.FindParentByAttr("role", "modal-overlay")  // *Element - ancestor search
```

### DOM Tree Manipulation

```go
el.InsertChild(child, index)
el.AppendChild(child)
el.Detach()                          // Remove from parent (keep handle)
el.Delete()                          // Delete element
el.Clone()                           // *Element
el.Swap(other)                       // Swap positions
```

### Size & Position

```go
el.ContentBox()    / el.ContentBoxSize()    // Rect / Size
el.PaddingBox()    / el.PaddingBoxSize()
el.BorderBox()     / el.BorderBoxSize()
el.MarginBox()     / el.MarginBoxSize()
```

### Other Operations

```go
el.Type()                            // string - tag name
el.GetElementUid()                   // uint32
el.Describe()                        // string - debug description
el.Update(restyle, restyleDeep, remeasure, remeasureDeep, render)
el.ScrollToView(toTop bool)
el.SetTimer(ms uint, timerId uintptr)
el.Capture() / el.ReleaseCapture()
el.ShowPopup(anchor, placement) / el.ShowPopupAt(x, y, animate) / el.HidePopup()
el.CallBehaviorMethod(methodId)      // e.g. DO_CLICK
el.SetEventRoot() / el.ResetEventRoot()
el.AttachHandler(handler, subscription)
```

### Per-Element Event Callbacks

```go
el.OnClick = func(owner *gohl.Element) bool { return true }
el.OnMouse = func(owner *gohl.Element, params *gohl.MouseParams) bool { return false }
el.OnButtonStateChanged = func(owner *gohl.Element, checked bool) bool { return false }
el.OnSelectionChanged = func(owner *gohl.Element, value string) bool { return false }
el.OnVisibleChange = func(owner *gohl.Element, visible bool) bool { return false }
el.OnValueChange = func(owner *gohl.Element, value string) bool { return false }
el.OnHyperlinkClick = func(owner *gohl.Element) bool { return true }
```

## UI Thread Dispatch

```go
// Dispatch: post task to UI thread from goroutine (recommended)
go func() {
    win.Dispatch(func() {
        el := win.GetElementById("status")
        if el != nil { el.SetText("done") }
    })
}()

// UpdateUI: batch UI updates (internally uses Dispatch)
win.UpdateUI(
    gohl.U{ID: "status", Action: "text", Value: "done"},
    gohl.U{ID: "progress", Action: "attr", Value: map[string]interface{}{"style": "width:80%"}},
    gohl.U{ID: "modal", Action: "show", Value: false},
)

// SetTimer: UI thread timer
win.SetTimer(1000, func() { /* every second */ })
```

### UpdateUI Actions

| Action | Value Type | Description |
|--------|-----------|-------------|
| `"text"` | `string` | SetText (empty string panics!) |
| `"html"` | `string` | SetHtml |
| `"value"` | `string` | SetValue |
| `"class"` | `string` | Set class attribute |
| `"addClass"` | `string` | Append class |
| `"removeClass"` | `string` | Remove class |
| `"show"` | `bool` | true=Show(), false=Hide() |
| `"hide"` | ignored | Hide() |
| `"attr"` | `map[string]interface{}` | Batch set attributes |
| `"style"` | `map[string]interface{}` | Batch set styles |
| `"enabled"` | `bool` | true=enable, false=disable |

## Storage (KV Store)

```go
store, err := gohl.NewStorage("MyApp")           // %APPDATA%/MyApp/storage.dat
store, err := gohl.NewStorageWithPath("C:/data.dat")  // Custom path

// Read/Write
store.Set("key", "value")
store.Set("count", 42)                           // Any gob-encodable type
val, ok := store.Get("key")                      // (interface{}, bool)
str, ok := store.GetString("key")                // (string, bool)
num, ok := store.GetInt("key")                   // (int64, bool)
b, ok := store.GetBool("key")                    // (bool, bool)

// TTL
store.SetWithTTL("temp", "data", 5*time.Minute)
ttl, ok := store.TTL("key")                      // (time.Duration, bool)
store.Expire("key", 10*time.Second)
store.Persist("key")                             // Remove expiration

// Delete & Check
store.Delete("key")
store.Exists("key")                              // bool
store.Keys()                                     // []string

// Batch
store.Clear()
store.CleanExpired()
store.Save() / store.Load() / store.Close()
store.FilePath()                                 // string
```

> Uses gob encoding, atomic write (temp file + rename). Auto-saves on each Set/Delete.

## TrayIcon

```go
tray := gohl.NewTrayIcon(gohl.TrayConfig{
    Icon: win.GetIcon(),
    Tip:  "App Name",
    UId:  1,  // optional, default 1
})

tray.Add(win.GetHwnd(), WM_TRAYMSG)    // bool
tray.ShowInfo("Title", "Content")       // bool
tray.ShowWarning("Title", "Content")    // bool
tray.ShowError("Title", "Content")      // bool
tray.ShowBalloon("Title", "Content", flags, timeout)
tray.SetIcon(newIcon)                   // bool
tray.SetTip("New tip")                  // bool
tray.IsAdded()                          // bool
tray.Remove()                           // bool
```

> Tray message: `WM_TRAYMSG = 0x0400 + 1`. Window wndProc handles left-click/double-click to restore.

## Resource Loading

```go
gohl.RegisterResourceLoader("embed", func(uri string) ([]byte, uint32, bool) {
    data := loadData(uri)  // uri: "embed://path/to/file"
    return data, gohl.GetResourceDataType(uri), true
})

// 回调拿到的 URI 末尾带斜杠：@font-face 里写 url('embed://iconfont.ttf')，
// 传进来是 "embed://iconfont.ttf/"。按后缀比名字会静默失配，先剥前缀再剥尾斜杠：
//   name := strings.TrimSuffix(strings.TrimPrefix(uri, "embed://"), "/")
// 字体不用自报类型：GetResourceDataType 对 .ttf 返回 0（AUTO），实测照常渲染。
//
// 想运行时换字体：gohl 的 defaultOnLoadData 按 URI 缓存字节（htmlayout_ui.go 里
// loadedResources[uri]），同一个 URI 只会交付第一次那份。所以每换一份就把
// @font-face 的 URI 换个序号，并用 gohl.LoadHtml(hwnd, doc, "") 整份文档重载。
//
// iconfont.cn 下载的 ttf 整支挂在 PUA（E000-F8FF），在 HTMLayout 里画出来是 '?'——
// 它按 GBK 解码，PUA 过不了。可用的码位是 56xx-59xx 这段 CJK（GBK 解得开，1024 个坑）。
// 这件事框架已经接手：resources.zip 里的 iconfont.new.ttf 是补好别名的那一份，页面里
// 一行 @font-face 就能用；自己下载的字体用 gohl.LoadIconfont 处理（见下节 Icon Fonts）。

// Built-in protocols:
// resources://filename  - auto-extracted resources.zip (fonts etc.)
// embed://filename      - user custom embed.zip
// file:///path          - local files

gohl.GetResourceDataType("image.png")  // uint32 - auto-detect by extension
// .html/.htm -> HLRT_DATA_HTML
// .css       -> HLRT_DATA_STYLE
// .js        -> HLRT_DATA_SCRIPT
// .png/.jpg/.gif/.bmp/.ico/.svg -> HLRT_DATA_IMAGE
```

## Icon Fonts (iconfont)

```go
// 内置那支（resources.zip 里的 iconfont.new.ttf）不需要写 Go 代码，页面里直接：
//   @font-face{ font-family:'iconfont'; src:url('resources://iconfont.new.ttf'); }
// 家族名就是字体自报的 'iconfont'；358 个可用码位，铺在 U+5600-U+5946 这段。

data, _ := gohl.ReadResource("iconfont.new.ttf") // 从内嵌 zip 直接取字节，不等释放到磁盘

f, _ := gohl.LoadIconfont(myTTF)   // 补 PUA 别名 + 换成唯一家族名 + 解出码位
uri := f.Serve("icon")             // 注册 scheme 并返回 URI："icon://iconfont-1.ttf"
html := "<style>" + f.CSS(uri) + "</style>"   // @font-face 一行
for _, g := range f.Glyphs {                   // 按码位升序；PUA 原件也在表里，画不出来
    if g.Codepoint >= 0x5600 && g.Codepoint <= 0x59FF { /* 页面上能用的 */ }
}
_ = f.Aliases                      // 新造了多少个 56xx 别名（本来就有别名的为 0）

gohl.AliasIconfont(myTTF)          // 只补别名、不改名：打包资源时用这条（tools/fontpack）
gohl.GlyphsOf(anyTTF)              // 只读码位表，不改字节（ttf/otf/ttc 都认）
gohl.FontFamily(anyTTF)            // 读自报家族名：@font-face 必须和它一字不差
gohl.BundledIconfont()             // = LoadIconfont(ReadResource("iconfont.new.ttf"))
gohl.ResourceRequested(uri)        // 引擎到底有没有来要过这个资源（排图标不出图用）
```

- 家族名必须唯一：GDI 按家族名挑 face，同名第二份注册了也不生效（实测仍回第一份的字形号），
  所以 `LoadIconfont` 会把名字换成 `<原名>-<序号>`；`f.Family` 就是 CSS 里要写的那个。
- URI 也必须唯一：`defaultOnLoadData` 按 URI 缓存字节，同一 URI 只会交付第一次那份。
  `Serve` 给的 URI 带上了唯一家族名，换字体后 `gohl.LoadHtml(hwnd, doc, "")` 整份重载。
- 想看/挑图标：`tools/iconpreview`（`go run ./tools/iconpreview`，可带一个 ttf 路径参数）。
  点格子即复制 HTML 写法，右下角能导出整段 Go 常量表。
- 换内置字体：把新字体放成 `testdata/iconfont.new.ttf` 后跑 `go run ./tools/fontpack`，
  它补好别名再写回 resources.zip，页面侧不用改。

## Built-in Behaviors

### tabs

```html
<div style="behavior: tabs">
    <div class="strip">
        <div panel="tab1" selected>Tab 1</div>
        <div panel="tab2">Tab 2</div>
    </div>
    <div name="tab1">Content 1</div>
    <div name="tab2">Content 2</div>
</div>
```

### light-box-dialog

```html
<div style="behavior: light-box-dialog" class="modal">
    <button role="ok-button">OK</button>
    <button role="cancel-button">Cancel</button>
</div>
```

Enter triggers `[role='ok-button']`, Escape triggers `[role='cancel-button']`.

### hyperlink

```html
<a style="behavior: hyperlink" href="...">Link</a>
```

Triggers `HYPERLINK_CLICK` via `win.OnHyperlinkClick` or `el.OnHyperlinkClick`.

## Utility Functions

```go
gohl.GetDpiScale()                       // float64 - DPI scale factor
gohl.LoadIconFromResource(id int)        // uintptr - icon from exe resource
gohl.ExtractIcon(filePath string, index int)  // uintptr - icon from file
gohl.GetDefaultIcon()                    // uintptr

gohl.SetOption(hwnd, option, value)      // bool
// HTMLAYOUT_SMOOTH_SCROLL = 1
// HTMLAYOUT_CONNECTION_TIMEOUT = 2
// HTMLAYOUT_HTTPS_ERROR = 3
// HTMLAYOUT_FONT_SMOOTHING = 4  (0=system, 1=none, 2=standard, 3=ClearType)
// HTMLAYOUT_ANIMATION_THREAD = 5
// HTMLAYOUT_TRANSPARENT_WINDOW = 6
```
