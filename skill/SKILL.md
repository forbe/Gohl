---
name: gohl-htmlayout
description: "Build Windows native desktop GUI apps with gohl + HTMLayout (Go backend + HTML/CSS frontend, no WebView). Use when creating/maintaining gohl desktop apps, handling HTMLayout layout quirks, or compiling exe with icon/manifest. Covers Window/Element/Storage/TrayIcon APIs, HTMLayout CSS differences from CSS3, and known pitfalls."
---

# gohl + HTMLayout Desktop Development

Go binding for HTMLayout engine (Sciter precursor). **Go backend + HTML/CSS frontend**, not WebView, no node/npm.

> This skill lives inside the gohl repo at `skill/` — the references below are relative to it.
> gohl source: `D:\codes\go\Gohl`
> Full API: `references/gohl-api.md` | CSS guide: `references/htmlayout-css.md` | Templates: `references/templates.md` | Official samples: `references/samples.md`
> Framework tools: `tools/iconpreview` (browse/copy iconfont glyphs), `tools/fontpack` (repack the bundled font into `resources.zip`)
> Official HTMLayout sample tree on disk: `D:\software\aardio\example\WebUI\HTMLayout\html_samples` (read before inventing CSS)

## Architecture

- HTMLayout renders UI with **CSS 2.1 + partial CSS3 + proprietary extensions** (NOT standard CSS3)
- Go <-> frontend via DOM API; all UI ops on **main thread** (gohl auto `runtime.LockOSThread()`)
- Recommended: `Frameless: true` + `Rounded: true`, self-draw titlebar with dot buttons
- Single-file deploy: `//go:embed` for HTML/resources zip + `rsrc` for icon/manifest
- gohl auto-extracts `resources.zip` to `%APPDATA%/gohl/`, served as `resources://`. It ships `alibaba_puhui.ttf`, `htmlayout.dll` and `iconfont.new.ttf` (alias-patched iconfont). Extraction is **per-file by size**, not all-or-nothing on a marker, so a file added in a newer gohl version still lands on an existing install. `gohl.ReadResource(name)` reads the bytes straight out of the embedded zip when you don't want to wait for the drop.
- **`resources://` / `embed://` are gohl-binding schemes**, resolved by its data-request callback — HTMLayout itself only knows relative paths, `res:`, `data:`, `theme:`
- Official sample tree is the ground truth for "does HTMLayout support X?": `D:\software\aardio\example\WebUI\HTMLayout\html_samples`. **Trap: `HTML32.htm` is a copy of the W3C HTML spec, not a feature doc** — use `cssmap.htm` for the CSS property list. Comments in that tree are GBK, so grep output shows garbled Chinese.

## Project Structure

```
myapp/
  main.go          # entry + window creation
  ui.go            # UI events + helpers
  index.html       # frontend (embedded or external)
  embed.zip        # bundled resources (images etc.)
  icon.ico         # app icon
  app.manifest     # DPI awareness
  build.bat        # compile script
  go.mod
```

## Quick Start

```go
var win *gohl.Window

func main() {
    win = gohl.NewWindow(gohl.WindowConfig{
        Title: "App", Width: 880, Height: 620,
        Frameless: true, Resize: true, Center: true,
        Rounded: true, CornerRadius: 10, Border: true,
        Icon: gohl.LoadIconFromResource(2),
    })
    win.LoadFile("index.html").Run()
}
```

## Critical Rules

1. **UI thread only** — use `win.Dispatch(fn)` or `win.UpdateUI(...)` from goroutines
2. **`SetText("")` and `SetHtml("")` panic** — the binding takes `&bytes[0]` on an empty slice (`element.go:817`). Pass `" "` for empty content; use `SetValue("")` for inputs
3. **`<input>` values** — must use `GetValue()`, not `ValueAsString()` or `Text()`
4. **`SetHtml()` content** — must be compact single-line (newlines become text nodes)
5. **No `linear-gradient()`** — use four-corner colors: `background-color: A B C D`
6. **No flexbox** — use `flow: horizontal/vertical` + `%%` / `*` units
7. **`behavior: clickable`** required on elements to trigger `OnButtonClick`
8. **`@font-face` must be first rule** in `<style>`
9. **`%` vs `%%` vs `*`** — `%` = parent-relative; `%%` = remaining space %; `*` = flex share
10. **No inline `<svg>`** — use as image resource only
11. **Charset must be `<meta http-equiv="Content-Type" content="text/html; charset=utf-8">`** — HTML5 `<meta charset="utf-8">` is NOT parsed; page falls back to single-byte codepage → every Chinese glyph in the static HTML becomes mojibake (`总` → `€`-ish). Symptom that identifies it: Go-injected text (`SetHtml`) renders fine while the hand-written HTML text is garbled.
12. **`@font-face` family name must equal the font's internal family name** — an arbitrary alias silently falls back to a serif default. The bundled `alibaba_puhui.ttf` is `'阿里巴巴普惠体 2.0 55 Regular'`, not `'Alibaba PuHuiTi'`.
13. **`min-width` defaults to `auto`** — a `flow` item refuses to shrink below its content, so per-row div "columns" never align. For any cross-row aligned column grid use `<table>`, or set `min-width: 0`.
14. **Ellipsis needs all three**: `white-space: nowrap; overflow-x: hidden; text-overflow: ellipsis;`
15. **No `z-index`** (0 hits in the whole official sample tree) — control stacking by DOM order or `position: fixed` overlays.
16. **Native controls (`<button>` `<input>` `<select>`) need no `behavior:`** — only div-as-widget needs `clickable`.
17. **`@set` scrollbar name is bare** — `vertical-scrollbar: small-v-scrollbar;`, no `@` prefix.
18. **No script** — gohl exposes no `Invoke`/`CallFunction`/`Eval`; CSSS! `xxx!:` properties and TIScript are unusable. Only plain-CSS `calc()` is available.
19. **Never put PUA codepoints (U+E000–U+F8FF) in the HTML text** — they are not GBK characters, so the HTML decode replaces each one with `U+003F` (`?`) before layout sees it; icon fonts then render as `?` even though the font itself loaded fine. Proof of which leg broke: read the DOM back (`win.GetElementById(id).Text()`) and compare runes with what the source file holds. **gohl now does the aliasing for you**: `resources://iconfont.new.ttf` is a cmap-patched iconfont whose glyphs also answer to CJK codepoints (`U+E666` ↔ `U+5666`, same glyph id), so one `@font-face` line and CJK-alias text is all a page needs. Your own iconfont.cn download: `gohl.LoadIconfont(data)` patches + renames it, `.Serve(scheme)`/`.CSS(uri)` hand you the wiring, and `tools/iconpreview` is the picker. See "Icon Fonts" in `references/gohl-api.md`.
20. **Don't tear the window down by hand** — `DestroyWindow` skips everything the binding puts in `WM_CLOSE` (sets `Window.closing`, stops `HTMLAYOUT_ANIMATION_THREAD`, detaches notify/event handlers, clears `loadedResources`). Queued HTMLayout messages then reach `wndProc` with `closing == false`, get forwarded to a detached layout, and the process dies with `Exception 0xc0000005` at exit (stack: `wndProc` → `ProcNoDefault(msg=0x4ae)` → `HTMLayoutProcND`). `win.Close()` posts `WM_CLOSE` and is safe (it used to be the raw `DestroyWindow`); `-gohl-close` routes through it too. Post, not send — a close usually runs inside a button callback.
21. **Ship `app.manifest` with PerMonitorV2 or every glyph is blurry** — without a DPI declaration the process is DPI-unaware: HTMLayout lays out and rasterizes at 96dpi, then DWM bitmap-stretches the whole window (on a 150% panel that is a 1.5× interpolation of everything — text, 1dip hairlines, icon fonts). Nothing in CSS or in `FONT_SMOOTHING` fixes it; it looks like "bad font rendering" but is pure scaling. gohl does NOT bring the manifest into your project — its own comment says 「DPI 感知由 manifest 文件设置 (PerMonitorV2)」 and the template only exists at `examples/app.manifest` inside the module. gohl then multiplies `config.Width/Height` by the DPI scale (`htmlayout_ui.go` `GetDpiScale`), so a 1200×780 window becomes 1800×1170 at 150% — intended, but check it still fits the screen. Verify embedding without a linker dump: `grep -ac PerMonitorV2 myapp.exe` must be ≥ 1.

## Window Special Attributes (HTML)

```html
<div -gohl-drag>...</div>       <!-- draggable titlebar -->
<button -gohl-min>─</button>    <!-- minimize -->
<button -gohl-max>□</button>    <!-- maximize/restore -->
<button -gohl-close>✕</button>  <!-- close -->
```

> `-gohl-close` routes through `win.Close()`, which posts `WM_CLOSE` (rule 20) — safe to use.

## Key APIs (summary)

```go
// Window control
win.Run() / win.Close() / win.Show() / win.Hide()   // Close posts WM_CLOSE — safe (rule 20)
win.Minimize() / win.Maximize() / win.Restore()
win.GetElementById(id) / win.GetRootElement()

// Element
el.Attr("id") / el.SetAttr("class", "active")
el.Text() / el.SetText("text") / el.SetHtml("<b>html</b>")
el.GetValue() / el.SetValue("val")
el.Show() / el.Hide() / el.AddClass("x") / el.RemoveClass("x")
el.FindParentByAttr("role", "modal-overlay")

// UI thread
win.Dispatch(func() { ... })
win.UpdateUI(gohl.U{ID: "x", Action: "text", Value: "hello"})

// Storage
store, _ := gohl.NewStorage("AppName")
store.Set("key", "value") / store.GetString("key")

// Tray
tray := gohl.NewTrayIcon(gohl.TrayConfig{Icon: win.GetIcon(), Tip: "App"})
tray.Add(win.GetHwnd(), WM_TRAYMSG)
```

## Event Routing Pattern

```go
win.OnButtonClick = func(elem *gohl.Element, params *gohl.BehaviorEventParams) bool {
    role, _ := elem.Attr("role")
    id, _ := elem.Attr("id")
    switch role {
    case "copy":     // handle copy
    case "show-modal": // handle modal
    }
    switch {
    case id == "btn-create": // handle create
    case strings.HasPrefix(id, "card-"): // handle dynamic cards
    }
    return false
}
```

## Default Font

```html
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<style>
    /* family MUST be the font's internal name — an alias falls back to serif */
    @font-face {
        font-family: '阿里巴巴普惠体 2.0 55 Regular';
        src: url('resources://alibaba_puhui.ttf');
    }
    * { font-family: '阿里巴巴普惠体 2.0 55 Regular'; font-size: 14dip; }
</style>
</head>
```

> No `<!DOCTYPE html>` — HTMLayout ignores it and the HTML5 `<meta charset>` shortcut that comes with it.

## Icon Fonts

```html
<style>
    @font-face { font-family: 'iconfont'; src: url('resources://iconfont.new.ttf'); }
    .icon { font-family: 'iconfont'; }
</style>
<span class="icon">&#x5666;</span>   <!-- CJK-alias codepoint, NOT U+E666 -->
```

Nothing to write in Go for the bundled font. Your own iconfont.cn download:

```go
f, _ := gohl.LoadIconfont(raw)      // patch cmap + give it a unique family
uri := f.Serve("icon")              // register the loader, returns the URI to reference
style := f.CSS(uri)                 // the @font-face line
for _, g := range f.Glyphs { ... }  // codepoints + post-table names, ascending
```

Pick glyphs visually with `go run ./tools/iconpreview` (click a cell to copy its HTML, button to export a Go const table); rebake the bundled font with `go run ./tools/fontpack`. Details: "Icon Fonts (iconfont)" in `references/gohl-api.md`.

## Compile

`app.manifest` — required, this is what makes text crisp (rule 21). Pure ASCII except the `name`:

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity version="1.0.0.0" processorArchitecture="*" name="MyApp" type="win32"/>
  <application xmlns="urn:schemas-microsoft-com:asm.v3">
    <windowsSettings>
      <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware>
      <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2, PerMonitor</dpiAwareness>
    </windowsSettings>
  </application>
</assembly>
```

```bat
@echo off
set PATH=D:\msys64\mingw64\bin;%PATH%
go mod tidy
go install github.com/akavel/rsrc@latest
rsrc -ico icon.ico -manifest app.manifest -o icon.syso
set CGO_ENABLED=1
go build -ldflags="-H windowsgui" -o myapp.exe
```

`.syso` is auto-linked by `go build` — but **only if it is regenerated**; a stale `icon.syso` silently keeps the old icon/manifest, so always re-run `rsrc`. `-H windowsgui` hides the console.

`CGO_ENABLED=1` is mandatory: gohl links through gcc, and `CGO_ENABLED=0` fails in `ld` with a wall of `undefined reference to '_cgo_yield' / 'x_cgo_init' / 'x_crosscall2_ptr'` — that is not a missing symbol, it is the cgo runtime being skipped.

If the Windows user name is non-ASCII (`C:\Users\江佳\...`), point the temp dir at an ASCII path or `ld.exe` reports `cannot open output file C:\Users\????\...\a.out.exe: No such file or directory`: `set TMP=D:\codes\gotmp` and `set TEMP=D:\codes\gotmp`.

After building, confirm the manifest actually went in before blaming the font: `grep -ac PerMonitorV2 myapp.exe`.

## Reference Files

- **`references/gohl-api.md`** — Complete Window/Element/Storage/TrayIcon/Behavior API
- **`references/htmlayout-css.md`** — HTMLayout CSS differences: encoding, fonts, **font clarity (DPI manifest first, then `dip`, font size floor, `FONT_SMOOTHING`)**, units, flow layout, `<table>`, backgrounds, transitions, selectors
- **`references/templates.md`** — Reusable HTML/Go patterns: titlebar, modal, toast, sidebar, page switching, progress bar, helper functions
- **`references/samples.md`** — Index of the official HTMLayout sample tree: which sample proves which layout claim, and the list of things that appear **nowhere** in it (`z-index`, `<col>`, `table-layout`, `@keyframes`, `placeholder`, `position:sticky` …). Read it before answering "does HTMLayout support X?".

## Framework Tools

- **`tools/iconpreview`** — browse + copy the iconfont glyphs (`go run ./tools/iconpreview [some.ttf]`; no argument = the bundled font). Clicking a cell copies its HTML usage; "导出 Go 常量" puts a whole const table on the clipboard.
- **`tools/fontpack`** — `go run ./tools/fontpack [src.ttf] [zip]`: alias-patches the font and rewrites that one entry in `resources.zip` (other entries byte-identical), so a new bundled iconfont needs zero page-side changes.
