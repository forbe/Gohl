package main

// iconpreview 是 gohl 自带的图标字体浏览器：看一支 ttf 里到底有哪些码位、点一下就复制
// 页面里能用的写法。
//
//	go run ./tools/iconpreview            # 打开就停在内置的 iconfont 上
//	./iconpreview.exe C:\path\to\x.ttf    # 停在指定字体上
//
// 内置那份默认预览的就是打进 resources.zip 的 iconfont.new.ttf（已补好 56xx 别名）：
// 预览器显示的码位必须是真实交付的那份字体，不然选出来的字符到了界面上是 ?。

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/forbe/gohl"
)

//go:embed index.html
var uiHTML string

const (
	appTitle = "图标字体预览"
	// maxCells 是一次塞进 DOM 的格子上限：随便一支中文字体就两万个码位，全写进去
	// 引擎会卡死，所以只画前面这些，剩下的靠过滤框缩小范围。
	maxCells = 1500
)

type app struct {
	win       *gohl.Window
	font      *gohl.Iconfont // 当前交付给引擎的那份（Data/Family/Glyphs 都在里面）
	glyphs    []gohl.Glyph   // 界面上排出来的那批，可用码位在前
	fontURI   string         // HTML 里 @font-face 写的 URI
	fontLabel string         // 界面上写的字体来源
	startFont string         // 命令行给的第一份字体，界面开在它上面（对话框没法自动化）

	sel   rune
	size  int
	all   bool // false = 只列非 PUA 的码位
	query string

	toastAt time.Time
}

// doc 把 index.html 里的 @@URI@@ / @@FAMILY@@ 换成当前这一版的 URI 和家族名。
func (a *app) doc() string {
	s := strings.Replace(uiHTML, "@@URI@@", a.fontURI, 1)
	return strings.ReplaceAll(s, "@@FAMILY@@", a.font.Family)
}

// serveFont 把当前这份字体交给引擎：走框架的 Iconfont.Serve。它给的 URI 里带唯一家族名，
// 引擎按 URI 缓存资源，名字不唯一就会把上一份端出来。
func (a *app) serveFont() string {
	return a.font.Serve("font")
}

func sortGlyphs(list []gohl.Glyph) {
	sort.Slice(list, func(i, j int) bool {
		x, y := list[i].Codepoint, list[j].Codepoint
		if (x >= 0xE000) != (y >= 0xE000) {
			return x < 0xE000
		}
		return x < y
	})
}

func main() {
	// 内置这份也走同一条交付路径：它一样自报 iconfont，不改名就会和后面打开的字体撞车。
	bundled, err := gohl.BundledIconfont()
	if err != nil {
		errBox("读内置 iconfont 失败: " + err.Error())
		return
	}
	a := &app{font: bundled, glyphs: bundled.Glyphs, size: 32,
		fontLabel: "内置 iconfont.new.ttf", fontURI: "font://" + bundled.Family + ".ttf"}
	sortGlyphs(a.glyphs)
	if len(os.Args) > 1 {
		a.startFont = os.Args[1] // 直接开在这份字体上，省掉点对话框那一步
	}

	var win *gohl.Window
	win = gohl.NewWindow(gohl.WindowConfig{
		Title:     appTitle,
		Width:     1020,
		Height:    760,
		Frameless: true,
		Resize:    true,
		Center:    true,
		Border:    true,
		Rounded:   true,
		Handler: gohl.NotifyHandler{
			OnDocumentComplete: func() uintptr {
				gohl.SetOption(uint32(win.GetHwnd()), gohl.HTMLAYOUT_FONT_SMOOTHING, 3)
				a.boot()
				return win.GetHwnd()
			},
		},
	})
	a.win = win
	a.serveFont()

	win.OnButtonClick = func(elem *gohl.Element) bool {
		id, _ := elem.Attr("id")
		a.route(id)
		return true
	}
	win.OnHyperlinkClick = func(elem *gohl.Element) bool {
		return a.routeLink(elem)
	}
	win.OnValueChange = func(elem *gohl.Element, value string) bool {
		id, _ := elem.Attr("id")
		if id == "q" {
			a.query = strings.TrimSpace(value)
			a.renderGrid()
		}
		return true
	}

	win.SetHtml(a.doc()).Run()
}

// boot 在每次文档加载完成后跑一次（换字体是重载文档，所以它会再进来）。
func (a *app) boot() {
	if a.startFont != "" {
		path := a.startFont
		a.startFont = "" // 先清掉：换字体要重载文档，boot 会再进来一次
		raw, err := os.ReadFile(path)
		if err != nil {
			a.toast("读不到 " + path + ": " + err.Error())
		} else {
			a.switchFont(raw, filepath.Base(path))
			return
		}
	}
	a.marks()
	a.fontStat()
	a.renderGrid()
	list := a.usable()
	if !inList(list, a.sel) && len(list) > 0 {
		a.sel = list[0].Codepoint
	}
	if len(list) > 0 {
		a.selectGlyph(a.sel)
	}
	// 字体是懒加载的，文档加载完那一刻可能还没来要，所以过一会儿再报一次，
	// 免得把"还没要"误报成"不会要"。
	time.AfterFunc(1500*time.Millisecond, func() { a.post(a.fontStat) })
}

func inList(list []gohl.Glyph, r rune) bool {
	for _, g := range list {
		if g.Codepoint == r {
			return true
		}
	}
	return false
}

func (a *app) fontStat() {
	// 引擎有没有来要过字体，直接写在界面上：图标显示成本字时，这一格先替我把范围
	// 缩小到"没要过"还是"要过没认"。
	st := "没来取（@font-face 或 scheme 没生效）"
	if gohl.ResourceRequested(a.fontURI) {
		st = "引擎已取到"
	}
	a.setText("fontstat", "字体 "+a.fontLabel+" · "+st)
	hint := "码位取自这份字体自己的 cmap 表；HTMLayout 走 GBK 解码，PUA（E000-F8FF）画不出来"
	if a.font.Aliases > 0 {
		hint = fmt.Sprintf("码位取自这份字体自己的 cmap 表；它挂在 PUA 的 %d 个字形已自动补了 56xx 别名，页面上用的是别名那批码位", a.font.Aliases)
	}
	a.setText("fonthint", hint)
}

func (a *app) el(id string) *gohl.Element {
	if a.win == nil {
		return nil
	}
	e := a.win.GetElementById(id)
	if e == nil || !e.IsValid() {
		return nil
	}
	return e
}

func (a *app) setText(id, s string) {
	if e := a.el(id); e != nil {
		if s == "" {
			s = " "
		}
		e.SetText(s)
	}
}

func (a *app) post(fn func()) {
	if a.win == nil {
		return
	}
	a.win.Dispatch(fn)
}

// usable 是当前该列出来的码位：过滤 + 是否含 PUA。
func (a *app) usable() []gohl.Glyph {
	out := make([]gohl.Glyph, 0, len(a.glyphs))
	q := strings.ToLower(a.query)
	for _, g := range a.glyphs {
		if !a.all && g.Codepoint >= 0xE000 {
			continue
		}
		if q != "" {
			hex := fmt.Sprintf("%04x", g.Codepoint)
			if !strings.Contains(strings.ToLower(g.Name), q) && !strings.Contains(hex, q) {
				continue
			}
		}
		out = append(out, g)
	}
	return out
}

// renderGrid 把当前列表拼成一格表格。HTMLayout 的 SetHtml 里换行会变成文本节点，
// 所以整段必须压成一行。
func (a *app) renderGrid() {
	list := a.usable()
	shown := list
	truncated := false
	if len(shown) > maxCells {
		shown = shown[:maxCells]
		truncated = true
	}
	var b strings.Builder
	b.WriteString("<table><tr>")
	for i, g := range shown {
		if i > 0 && i%8 == 0 {
			b.WriteString("</tr><tr>")
		}
		cls := "cell"
		if g.Codepoint >= 0xE000 {
			cls += " pua"
		}
		if g.Codepoint == a.sel {
			cls += " sel"
		}
		name := g.Name
		if name == "" {
			name = "—"
		}
		fmt.Fprintf(&b, `<td><div class="%s" id="g-%04x"><span class="gl" style="font-size:%ddip">%s</span><span class="cp">U+%04X</span><span class="nm" title="%s">%s</span></div></td>`,
			cls, g.Codepoint, a.size, esc(string(g.Codepoint)), g.Codepoint, esc(name), esc(name))
	}
	for i := len(shown); i%8 != 0; i++ {
		b.WriteString("<td></td>")
	}
	b.WriteString("</tr></table>")
	if len(shown) == 0 {
		b.Reset()
		b.WriteString(`<div style="padding:20dip;color:#8b909a">没有匹配的码位 — 换个关键词，或点上面的"全部（含 PUA）"</div>`)
	}
	if e := a.el("grid"); e != nil {
		e.SetHtml(b.String())
		log.Printf("[grid] 已写入 %d 格，%d 字节", len(shown), b.Len())
	} else {
		log.Printf("[grid] 找不到 #grid，列表 %d 格没地方放", len(shown))
	}
	tail := ""
	if truncated {
		tail = fmt.Sprintf(" · 只画前 %d 格，用过滤缩小范围", maxCells)
	}
	a.setText("cnt", fmt.Sprintf("显示 %d / 共 %d（非 PUA %d · PUA %d）%s",
		len(shown), len(a.glyphs), a.count(false), a.count(true), tail))
}

func (a *app) count(pua bool) int {
	n := 0
	for _, g := range a.glyphs {
		if (g.Codepoint >= 0xE000) == pua {
			n++
		}
	}
	return n
}

// selectGlyph 更新下面那排可复制的写法。这里给全五种抄法，是因为它们各自只在一种
// 文件里对：HTML 里要原字符、Go 里要 \u 转义、CSS content 里要 \XXXX。
func (a *app) selectGlyph(r rune) {
	a.sel = r
	var gname string
	for _, g := range a.glyphs {
		if g.Codepoint == r {
			gname = g.Name
			break
		}
	}
	a.setText("big", string(r))
	a.setText("d-name", orDash(gname))
	a.setText("d-cp", fmt.Sprintf("U+%04X (十进制 %d)", r, r))
	a.setText("d-go", fmt.Sprintf(`"\u%04x"`, r))
	a.setText("d-css", fmt.Sprintf("content: \"\\%04x\";", r))
	a.setText("d-html", htmlSnip(r))
	a.setText("d-char", string(r))
	a.renderGrid()
}

// htmlSnip 是往 .ico 那一类标签里塞的写法，也是点格子时直接进剪贴板的内容。
func htmlSnip(r rune) string {
	return fmt.Sprintf("<span class=\"ico\">%s</span>", string(r))
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func (a *app) route(id string) {
	switch {
	case id == "win-close":
		a.win.Close()
	case id == "seg-usable":
		a.all = false
		a.marks()
		a.renderGrid()
	case id == "seg-all":
		a.all = true
		a.marks()
		a.renderGrid()
	case strings.HasPrefix(id, "sz-"):
		var n int
		fmt.Sscanf(id, "sz-%d", &n)
		if n > 0 {
			a.size = n
			a.marks()
			a.renderGrid()
		}
	case strings.HasPrefix(id, "g-"):
		var v uint64
		fmt.Sscanf(id, "g-%x", &v)
		if v != 0 {
			r := rune(v)
			a.selectGlyph(r)
			a.copyText(htmlSnip(r))
		}
	case id == "open-font":
		a.openFont()
	case id == "font-reset":
		a.resetFont()
	case id == "exp-go":
		a.exportGo()
	}
}

// copyText 直接写剪贴板并回报内容，供点格子这种"我就要这一段"的动作用。
func (a *app) copyText(txt string) {
	if err := setClipboard(txt); err != nil {
		a.toast("复制失败: " + err.Error())
		return
	}
	a.toast("已复制：" + txt)
}

// switchFont 换一份字体来看：解析失败就报出来、界面保持原样；成功则整份文档重载，
// 因为 @font-face 只在文档解析时读一次，改 DOM 里的样式表不顶用。
func (a *app) switchFont(raw []byte, label string) {
	f, err := gohl.LoadIconfont(raw)
	if err != nil {
		a.toast("处理字体失败，按原字体预览: " + err.Error())
		a.showRaw(raw, label)
		return
	}
	if f.Aliases > 0 {
		label += fmt.Sprintf(" · 已造 %d 个 56xx 别名", f.Aliases)
	}
	a.font, a.fontLabel = f, label
	a.glyphs = f.Glyphs
	sortGlyphs(a.glyphs)
	a.reload()
}

// resetFont 回到内置那一份：重新从 resources.zip 读，名字也跟着换一号，
// 免得引擎还按上一份的家族名挑 face。
func (a *app) resetFont() {
	f, err := gohl.BundledIconfont()
	if err != nil {
		a.toast("读内置字体失败: " + err.Error())
		return
	}
	a.font, a.fontLabel = f, "内置 iconfont.new.ttf"
	a.glyphs = f.Glyphs
	sortGlyphs(a.glyphs)
	a.reload()
}

// showRaw 是处理失败时的退路：按字体自己报的家族名交付，界面说清楚用的哪个名字。
func (a *app) showRaw(raw []byte, label string) {
	family := label
	if f, err := gohl.FontFamily(raw); err == nil {
		family = f
	}
	a.font = &gohl.Iconfont{Family: family, Data: raw}
	if list, err := gohl.GlyphsOf(raw); err == nil {
		a.glyphs = list
		sortGlyphs(a.glyphs)
	}
	a.reload()
}

func (a *app) reload() {
	log.Printf("[switch] %s: 家族名=%s 码位 %d 个（非 PUA %d）别名 %d 个",
		a.fontLabel, a.font.Family, len(a.glyphs), a.count(false), a.font.Aliases)
	// 外来图标字体常常整支挂在 PUA，"非 PUA"视图会是空的，这时直接切到全部。
	if a.count(false) == 0 {
		a.all = true
	}
	a.sel = 0
	a.fontURI = a.serveFont()
	if err := gohl.LoadHtml(uint32(a.win.GetHwnd()), []byte(a.doc()), ""); err != nil {
		a.toast("重载文档失败: " + err.Error())
	}
}

func (a *app) openFont() {
	path, ok := runOpenFontDialog(a.win.GetHwnd())
	if !ok {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		a.toast("读文件失败: " + err.Error())
		return
	}
	a.switchFont(data, filepath.Base(path))
}

// marks 只改那几个开关按钮的灰态，不动别的：HTMLayout 没有 :not()，选中态就用 class 手写。
func (a *app) marks() {
	on := func(id string, ok bool) {
		e := a.el(id)
		if e == nil {
			return
		}
		if ok {
			e.RemoveClass("off")
		} else {
			e.AddClass("off")
		}
	}
	on("seg-usable", !a.all)
	on("seg-all", a.all)
	on("sz-24", a.size != 24)
	on("sz-32", a.size != 32)
	on("sz-48", a.size != 48)
}

// exportGo 把当前列出来的这批码位导成一份常量表贴进剪贴板。名字取自 post 表，非法
// 字符换成下划线，重名的加序号——这是给 ui.go 当起手式的，不是让它自动写文件。
func (a *app) exportGo() {
	list := a.usable()
	if len(list) > maxCells {
		list = list[:maxCells]
	}
	seen := map[string]int{}
	var b strings.Builder
	n := 0
	b.WriteString("// 码位；名字来自字体 post 表，可直接改成业务名。\nconst (\n")
	for _, g := range list {
		base := ident(g.Name)
		if base == "" {
			base = fmt.Sprintf("Rune%04X", g.Codepoint)
		}
		key := base
		if c := seen[base]; c > 0 {
			key = fmt.Sprintf("%s_%d", base, c)
		}
		seen[base]++
		n++
		fmt.Fprintf(&b, "\t%s = \"\\u%04x\" // U+%04X\n", key, g.Codepoint, g.Codepoint)
	}
	b.WriteString(")")
	out := b.String()
	if err := setClipboard(out); err != nil {
		a.toast("导出失败: " + err.Error())
		return
	}
	a.toast(fmt.Sprintf("已复制 %d 个常量到剪贴板", n))
}

func ident(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return ""
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "_" + out
	}
	return out
}

func (a *app) routeLink(el *gohl.Element) bool {
	if el == nil {
		return false
	}
	role, _ := el.Attr("role")
	if role != "copy" {
		return false
	}
	target, _ := el.Attr("target")
	a.copyToClipboard(target)
	return true
}

func (a *app) copyToClipboard(id string) {
	if id == "" {
		a.toast("复制链接没写 target")
		return
	}
	e := a.el(id)
	if e == nil {
		a.toast("找不到要复制的内容：" + id)
		return
	}
	txt := strings.TrimSpace(e.Text())
	if txt == "" || txt == "—" {
		a.toast("这一格还没有内容可复制")
		return
	}
	if err := setClipboard(txt); err != nil {
		a.toast("复制失败: " + err.Error())
		return
	}
	a.toast("已复制：" + txt)
}

func (a *app) toast(s string) {
	a.toastAt = time.Now()
	a.setText("toast", s)
	time.AfterFunc(2600*time.Millisecond, func() {
		a.post(func() {
			if time.Since(a.toastAt) >= 2500*time.Millisecond {
				a.setText("toast", " ")
			}
		})
	})
}

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

// 剪贴板：与 gohl-ui 同一套 Win32 走法。CF_UNICODETEXT，GlobalAlloc 交出去后不自由释放。
var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modNtdll    = syscall.NewLazyDLL("ntdll.dll")

	procOpenClipboard  = modUser32.NewProc("OpenClipboard")
	procCloseClipboard = modUser32.NewProc("CloseClipboard")
	procEmptyClipboard = modUser32.NewProc("EmptyClipboard")
	procSetClipData    = modUser32.NewProc("SetClipboardData")
	procGlobalAlloc    = modKernel32.NewProc("GlobalAlloc")
	procGlobalFree     = modKernel32.NewProc("GlobalFree")
	procGlobalLock     = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock   = modKernel32.NewProc("GlobalUnlock")
	procMoveMemory     = modNtdll.NewProc("RtlMoveMemory")
	procMessageBox     = modUser32.NewProc("MessageBoxW")
)

func setClipboard(text string) error {
	const (
		gmemMoveable  = 0x0002
		gmemZero      = 0x0040
		cfUnicodeText = 13
	)
	u := append(utf16.Encode([]rune(text)), 0)
	size := uintptr(len(u) * 2)

	var opened bool
	for i := 0; i < 3; i++ {
		if r, _, _ := procOpenClipboard.Call(0); r != 0 {
			opened = true
			break
		}
		time.Sleep(60 * time.Millisecond)
	}
	if !opened {
		return fmt.Errorf("剪贴板被其它程序占用，稍等一下再复制")
	}
	defer procCloseClipboard.Call()

	h, _, e := procGlobalAlloc.Call(gmemMoveable|gmemZero, size)
	if h == 0 {
		return fmt.Errorf("申请剪贴板内存失败: %v", e)
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("锁定剪贴板内存失败")
	}
	procMoveMemory.Call(p, uintptr(unsafe.Pointer(&u[0])), size)
	procGlobalUnlock.Call(h)

	if r, _, _ := procEmptyClipboard.Call(); r == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("清空剪贴板失败")
	}
	if r, _, _ := procSetClipData.Call(cfUnicodeText, h); r == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("写入剪贴板失败")
	}
	return nil
}

func errBox(msg string) {
	p, _ := syscall.UTF16PtrFromString(msg)
	t, _ := syscall.UTF16PtrFromString(appTitle)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(t)), 0x10)
	os.Exit(1)
}

// 打开字体文件：gohl 没有文件对话框，直接调 comdlg32!GetOpenFileNameW。
var (
	modComdlg32           = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW  = modComdlg32.NewProc("GetOpenFileNameW")
	openExplorerFlags     = uint32(0x00080000) // OFN_EXPLORER
	openMustExistFlags    = uint32(0x00001000) // OFN_FILEMUSTEXIST
	openPathExistFlags    = uint32(0x00000800) // OFN_PATHMUSTEXIST
	openHideReadonlyFlags = uint32(0x00000004) // OFN_HIDEREADONLY
)

type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFileIndex        uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
}

// utf16z 把若干串拼成 "a\0b\0\0" 这种双终止的宽字符块（过滤器要这个形状）。
func utf16z(parts ...string) []uint16 {
	var b []uint16
	for _, p := range parts {
		b = append(b, utf16.Encode([]rune(p))...)
		b = append(b, 0)
	}
	return append(b, 0)
}

func runOpenFontDialog(owner uintptr) (string, bool) {
	fileBuf := make([]uint16, 4096)
	titleBuf := make([]uint16, 260)
	filter := utf16z("字体文件 (*.ttf;*.otf;*.ttc)", "*.ttf;*.otf;*.ttc", "所有文件 (*.*)", "*.*")
	title, _ := syscall.UTF16PtrFromString("选择要预览的字体")
	defExt, _ := syscall.UTF16PtrFromString("ttf")

	ofn := openFileNameW{
		hwndOwner:      owner,
		lpstrFilter:    &filter[0],
		lpstrFile:      &fileBuf[0],
		nMaxFile:       uint32(len(fileBuf)),
		lpstrFileTitle: &titleBuf[0],
		nMaxFileTitle:  uint32(len(titleBuf)),
		lpstrTitle:     title,
		lpstrDefExt:    defExt,
		flags:          openExplorerFlags | openMustExistFlags | openPathExistFlags | openHideReadonlyFlags,
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))

	r, _, e := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	runtime.KeepAlive(ofn)
	runtime.KeepAlive(filter)
	runtime.KeepAlive(fileBuf)
	if r == 0 {
		if e != syscall.Errno(0) {
			log.Printf("[dialog] GetOpenFileNameW 失败: %v", e)
		}
		return "", false // 取消也是 0，且 last error 为 0
	}
	return syscall.UTF16ToString(fileBuf), true
}
