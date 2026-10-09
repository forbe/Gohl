# Gohl

Go 语言 HTMLayout 绑定库，用于构建基于 HTML/CSS 的现代桌面应用程序界面(支持win7、无边框窗口、圆角窗口等)。

## 简介

Gohl 是 [HTMLayout](https://terrainformatica.com/htmlayout/) 引擎的 Go 语言封装，让你能够使用 HTML 和 CSS 来设计和渲染桌面应用程序的用户界面。它提供了一套完整的 DOM 操作 API 和事件处理机制，支持创建无边框、圆角等现代化窗口样式。

## 特性

- **HTML/CSS 渲染** - 使用标准的 HTML 和 CSS 构建界面
- **No CGO** - 完全的syscall调用，高性能，不挑编译环境
- **独立编译** 内置HTMLayout.dll、阿里巴巴普惠体、iconfont图标字体(自包含于resources.zip)，编译会独立生成exe文件，无需显式依赖HTMLayout.dll
- **无边框窗口** - 支持自定义标题栏和窗口控制
- **圆角窗口** - 支持设置窗口圆角半径
- **内置 Behaviors** - 提供 tabs、light-box-dialog、hyperlink 等常用组件行为
- **DOM 操作** - 完整的元素选择、属性操作、样式修改等 API
- **事件系统** - 支持鼠标、键盘、焦点、自定义事件等
- **资源加载** - 支持从文件或内存加载 HTML 和资源
- **内置图标字体** - resources.zip 带一支可直接使用的 iconfont，详见「图标字体」
- **定时器** - 支持一次性定时器回调

## 安装

```bash
go get github.com/forbe/gohl
```

> import 路径是小写的 `github.com/forbe/gohl`（模块名如此，与仓库大小写不同）。

## 快速开始

```go
package main

import (
    "github.com/forbe/gohl"
)

func main() {
    gw := gohl.NewWindow(gohl.WindowConfig{
        Title:        "My App",
        Width:        800,
        Height:       600,
        Frameless:    true,
        Resize:       true,
        Center:       true,
    })

    gw.OnClick = func(elem *gohl.Element) {
        id, _ := elem.Attr("id")
        switch id {
        case "close-btn":
            gw.Close()
        }
    }

    gw.LoadFile("index.html").Run()
}
```

## 窗口配置

```go
type WindowConfig struct {
    Title        string  // 窗口标题
    Width        int     // 窗口宽度
    Height       int     // 窗口高度
    ClassName    string  // 窗口类名
    Border       bool    // 是否显示边框
    Frameless    bool    // 无边框模式
    MaxBtn       bool    // 是否显示最大化按钮
    MinBtn       bool    // 是否显示最小化按钮
    Resize       bool    // 是否允许调整大小
    Center       bool    // 是否居中显示
    Icon         uintptr // 窗口图标
    Rounded      bool    // 是否圆角窗口
    CornerRadius int     // 圆角半径
    Handler: gohl.NotifyHandler{
		OnDocumentComplete: func() uintptr { ... }
	}
}
```

## HTML 窗口控制属性

在 HTML 中使用特殊属性来实现窗口控制：

```html
<div id="title-bar" -gohl-drag>
    <span>窗口标题</span>
    <button -gohl-min>-</button>
    <button -gohl-max>□</button>
    <button -gohl-close>×</button>
</div>
```

| 属性 | 功能 |
|------|------|
| `-gohl-drag` | 允许拖动窗口 |
| `-gohl-min` | 最小化窗口 |
| `-gohl-max` | 最大化/还原窗口 |
| `-gohl-close` | 关闭窗口 |

> `gw.Close()` 发的是 `WM_CLOSE`，会走完停动画线程、摘事件回调、清资源这几步；不要绕过它直接 `DestroyWindow`，
> 否则排队的 HTMLayout 消息会打到已脱钩的 layout 上，进程退出时 `0xc0000005`。

## 内置 Behaviors

### Tabs(未测试，谨慎)

```html
<div class="tabs" behavior="tabs">
    <div class="strip">
        <div panel="panel1" selected>Tab 1</div>
        <div panel="panel2">Tab 2</div>
    </div>
    <div class="panel" name="panel1">Content 1</div>
    <div class="panel" name="panel2">Content 2</div>
</div>
```

### Light-box Dialog

```html
<div class="modal" behavior="light-box-dialog">
    <div class="modal-header">标题</div>
    <div class="modal-body">内容</div>
    <div class="modal-footer">
        <button role="cancel-button">取消</button>
        <button role="ok-button">确定</button>
    </div>
</div>
```

### Hyperlink

```html
<a behavior="hyperlink">点击链接</a>
```

## DOM 操作

```go
root := gw.GetRootElement()

elem := root.GetElementById("my-element")

elem.SetText("Hello")
elem.SetHtml("<b>Hello</b>")
elem.SetAttr("class", "active")
elem.SetStyle("color", "red")
elem.Show()
elem.Hide()

value := elem.ValueAsString()
text := elem.Text()
```

## 事件处理

```go

### 直接绑定（优先级高）
var elem = gw.ElementById("some")
ele.OnClick = func(elem *gohl.Element) bool {}
ele.OnEditValueChanged = ...
ele.On.... = 

## 全局处理（优先级低）
gw.OnClick = func(elem *gohl.Element) {
    // 处理点击事件
}

gw.OnHyperlinkClick = func(elem *gohl.Element) {
    // 处理超链接点击
}

gw.OnEditValueChanged = func(elem *gohl.Element) {
    // 处理输入框值变化
}


```

## 自定义 Behavior

```go
gw.SetNotifyHandler(&gohl.NotifyHandler{
    Behaviors: map[string]*gohl.EventHandler{
        "my-behavior": {
            OnAttached: func(he gohl.HELEMENT) {
                // 元素附加时调用
            },
            OnMouse: func(he gohl.HELEMENT, params *gohl.MouseParams) bool {
                // 处理鼠标事件
                return false
            },
            OnBehaviorEvent: func(he gohl.HELEMENT, params *gohl.BehaviorEventParams) bool {
                // 处理行为事件
                return false
            },
        },
    },
})
```

## 定时器

```go
gw.SetTimer(2000, func() {
    // 2秒后执行
    log.Println("Timer fired!")
})
```

## 图标字体（iconfont）

HTMLayout 按 GBK 解码 HTML，iconfont.cn 那种挂在 PUA（U+E000-U+F8FF）上的字体在页面上只会画出一个 `?`。
框架把这件事接管了：resources.zip 里的 `iconfont.new.ttf` 是重写过 cmap 的那一份——每个图标除了 PUA
原件，还多一个 GBK 解得开的 CJK 别名（U+5600-U+5946，共 358 个可用码位）。**页面里一行 CSS 就能用，不用写 Go 代码**：

```html
<style>
    @font-face { font-family: 'iconfont'; src: url('resources://iconfont.new.ttf'); }
    .icon { font-family: 'iconfont'; font-size: 20dip; }
</style>

<span class="icon">&#x5666;</span>   <!-- 写 CJK 别名码位，别写 U+E666 -->
```

自己下载的字体交给框架处理（补别名 + 换成唯一家族名 + 解出码位表）：

```go
f, _ := gohl.LoadIconfont(raw)   // raw 是 ttf/otf 字节
uri := f.Serve("icon")           // 注册资源协议，返回 "icon://<家族名>.ttf"
html := "<style>" + f.CSS(uri) + "</style>"

for _, g := range f.Glyphs {     // 按码位升序，含 post 表里的图标名
    fmt.Printf("U+%04X %s\n", g.Codepoint, g.Name)
}
```

| API | 作用 |
|------|------|
| `gohl.LoadIconfont(data)` | 一步到位：补别名 + 唯一家族名 + 码位表，返回 `*Iconfont` |
| `(*Iconfont).Serve(scheme)` / `.CSS(uri)` | 注册资源回调、生成 `@font-face` |
| `gohl.BundledIconfont()` | 取内置那支（等价于 `LoadIconfont(ReadResource("iconfont.new.ttf"))`） |
| `gohl.AliasIconfont(data)` | 只补别名不改名，打包资源时用这条 |
| `gohl.GlyphsOf(data)` / `gohl.FontFamily(data)` | 只读码位表 / 只读自报家族名 |
| `gohl.ReadResource(name)` | 从内嵌 resources.zip 直接取字节，不等释放到磁盘 |
| `gohl.ResourceRequested(uri)` | 引擎到底来要过这个资源没有（图标不出图时先查这个） |

两个必须唯一的坑：GDI 按**家族名**挑 face（同名第二份注册了也不生效），gohl 按 **URI** 缓存字节（同一 URI 只交付第一次那份）。
`LoadIconfont` / `Serve` 已经把这两件事都处理掉了，换字体后整份文档重载即可（`gohl.LoadHtml(hwnd, doc, "")`）。

## 框架工具

```bash
go run ./tools/iconpreview            # 浏览、挑取图标；点格子复制 HTML 写法，一键导出 Go 常量表
go run ./tools/iconpreview my.ttf     # 换成看自己的字体
go run ./tools/fontpack               # 把 testdata/iconfont.new.ttf 补好别名后写回 resources.zip
```

`fontpack` 只改写目标那一条目，其余字节原样保留，产出是确定的——内置字体从此可复现。

## 开发文档（skill）

`skill/` 是随框架一起发布的 AI 技能包，内容是在真实项目里踩出来的规则，也可直接当文档读：

| 文件 | 内容 |
|------|------|
| `skill/SKILL.md` | 21 条关键规则（编码、布局、单位、DPI、图标字体、窗口关闭…）+ 快速上手 |
| `skill/references/gohl-api.md` | Window / Element / Storage / Tray / 资源加载 / 图标字体 完整 API |
| `skill/references/htmlayout-css.md` | HTMLayout 的 CSS 与标准 CSS3 的差异 |
| `skill/references/templates.md` | 标题栏、弹窗、Toast、侧边栏等可复用模板 |
| `skill/references/samples.md` | 官方示例索引：想知道"HTMLayout 支持 X 吗"先查这里 |

## 示例

查看 [examples/demo.go](examples/demo.go) 获取完整示例。

运行示例：

```bash
cd examples
build.bat
./demo.exe
```

![DEMO](https://github.com/forbe/Gohl/blob/main/ScreenShot.png?raw=true)
![DEMO](https://github.com/forbe/Gohl/blob/main/ScreenShot2.png?raw=true)
![ICONFONT](https://github.com/forbe/Gohl/blob/main/QQ%E6%88%AA%E5%9B%BE20261009175038.png?raw=true)



## 依赖

- Windows 操作系统
- Go 1.20
- HTMLayout DLL (htmlayout.dll已经打包在Resources.zip中，会自动释放)

resources.zip 释放到 `%APPDATA%\gohl`，通过 `resources://文件名` 引用。里面有 `htmlayout.dll`、
`alibaba_puhui.ttf`、`iconfont.new.ttf`。释放是**逐文件按大小比对**的，所以新版本新增的文件在老用户
机器上照样会落地，无需重装。

## 许可证

LGPL V3 License （遵循HTMLayout）

## 致谢

- [HTMLayout](https://terrainformatica.com/htmlayout/) - HTML/CSS 渲染引擎
