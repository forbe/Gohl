# Gohl

Go 语言 HTMLayout 绑定库，用于构建基于 HTML/CSS 的现代桌面应用程序界面。

## 简介

Gohl 是 [HTMLayout](https://terrainformatica.com/htmlayout/) 引擎的 Go 语言封装，让你能够使用 HTML 和 CSS 来设计和渲染桌面应用程序的用户界面。它提供了一套完整的 DOM 操作 API 和事件处理机制，支持创建无边框、圆角等现代化窗口样式。

## 特性

- **HTML/CSS 渲染** - 使用标准的 HTML 和 CSS 构建界面
- **无边框窗口** - 支持自定义标题栏和窗口控制
- **圆角窗口** - 支持设置窗口圆角半径
- **内置 Behaviors** - 提供 tabs、light-box-dialog、hyperlink 等常用组件行为
- **DOM 操作** - 完整的元素选择、属性操作、样式修改等 API
- **事件系统** - 支持鼠标、键盘、焦点、自定义事件等
- **资源加载** - 支持从文件或内存加载 HTML 和资源
- **定时器** - 支持一次性定时器回调

## 安装

```bash
go get github.com/yourusername/gohl
```

## 快速开始

```go
package main

import (
    "gohl"
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

    gw.SetNotifyHandler(&gohl.NotifyHandler{})

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

## 内置 Behaviors

### Tabs

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

## 示例

查看 [examples/demo.go](examples/demo.go) 获取完整示例。

运行示例：

```bash
cd examples
go build -o demo.exe demo.go
./demo.exe
```

## 依赖

- Windows 操作系统
- Go 1.25+
- HTMLayout DLL (htmlayout.dll)

## 许可证

MIT License

## 致谢

- [HTMLayout](https://terrainformatica.com/htmlayout/) - HTML/CSS 渲染引擎
