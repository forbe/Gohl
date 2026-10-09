# HTMLayout CSS Guide (Differences from CSS3)

HTMLayout uses **CSS 2.1 as baseline** + partial CSS3 + many proprietary extensions. This is NOT standard CSS3.

> 支持与否以官方样例为准，见 `references/samples.md`。浏览器 CSS 的经验在这里会静默失效——不报错，只是不生效。

## 文档编码（踩过的坑）

HTMLayout **不认** HTML5 的 `<meta charset="utf-8">` 简写。没有可识别的编码声明时，它按单字节码页解析 UTF-8 字节流，汉字被拆成 2–3 个字符（`总` = `E6 80 BB` 中间那字节渲染成 `€`），其余落在控制字符上→全是方块。

```html
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
```

- 必须用 `http-equiv` 全形式；`<!DOCTYPE html>` 一并去掉，与官方样例保持一致。
- 症状可用来定位路径：**静态文档里的中文是方块、而 Go 侧 `SetHtml` 注入的中文正常** → 就是这里，不是字体问题。

## 字体 @font-face

```css
@font-face { font-family: '阿里巴巴普惠体 2.0 55 Regular'; src: url('resources://alibaba_puhui.ttf'); }
* { font-family: '阿里巴巴普惠体 2.0 55 Regular'; font-size: 14dip; }
```

- **`font-family` 必须等于字体文件内部的 family 名**，不能随便取别名。官方唯一的样例 `@font-face/test.htm` 里 `NinjaLine` 恰好既是别名也是该 ttf 的内部名，所以它证明不了别名可用；实测写 `'Alibaba PuHuiTi'`（缺了版本段和字重段）**不生效**，界面回落到默认衬线体；改成 `'阿里巴巴普惠体 2.0 55 Regular'` 才生效（本仓 `gohl-ui/index.html` 已验证）。查真名：PowerShell `System.Windows.Media.GlyphTypeface` 的 `Win32FamilyNames`。
- 别名不匹配时没有任何报错，只能靠眼睛看（拉丁字母也会一起变成衬线体，这是最明显的信号）。
- `resources://` 与 `embed://` 是 **gohl 绑定层自己实现的 scheme**（`htmlayout_ui.go` 的 data-request 回调，从 `%APPDATA%\gohl\` 取文件），不是 HTMLayout 标准。HTMLayout 认的是相对路径、`res:`（引擎内建资源）、`data:`、`theme:`。
- 声明顺序：放在第一个 `<style>` 的最前面（惯例，官方样例未强制说明）。

## 字体清晰度（DPI 是第一因，别从 CSS 找）

- **先确认进程是不是 DPI-aware，再谈字体。** 没有内嵌 `app.manifest`（PerMonitorV2）就是 DPI-unaware：HTMLayout 按 96dpi 排版栅格化，DWM 再把整窗位图拉伸（150% 屏就是 1.5 倍插值）。特征是**整体一起糊**——文字、`1dip` 边框（变成两根灰线）、iconfont 字形全都不锐，这时 CSS 和 `FONT_SMOOTHING` 怎么调都没用。查 exe 有没有带声明，不用 dump 资源：`grep -ac PerMonitorV2 myapp.exe`（0 = 没带）。gohl 不会替项目带 manifest，模板只在模块的 `examples/app.manifest` 里；它自己注释写着「DPI 感知由 manifest 文件设置 (PerMonitorV2)」。
- **manifest 修好后窗口会变大**，这是 gohl 的既有逻辑不是 bug：建窗时把 `config.Width/Height` 乘了 `GetDpiScale()`（`htmlayout_ui.go:508`），150% 下 1200×780 → 1800×1170。要原来的观感就把 config 的宽高按逻辑尺寸调小，内容是流式的会自己 reflow。
- **单位一律 `dip`，不要混 `px`。** `px` 在 DPI-aware 后是物理像素，150% 屏上的 `1px` 分隔线等于 0.67dip，比设计稿细；`2px`/`1px` 这类残留要换成 `dip`。
- **小字号有下限**：11dip 在 150% 下只有 16.5 物理像素高，12dip 是 18 —— 汉字笔画已经挤在一起了。正文用 13~14dip，说明/次要信息最低 12dip。普惠体 55 Regular 的笔画比微软雅黑细，深色底 + 浅字时"薄"会被看成"虚"，这种情况换字重或加大字号，比调平滑有效。
- **`HTMLAYOUT_FONT_SMOOTHING`（选项号 4）**：`0` 系统默认 / `1` 不平滑 / `2` 标准灰度 / `3` ClearType（gohl 的注释原文）。追求锐利用 `3`；嫌它有彩色边、觉得没对准像素就 A/B 成 `2`（深色主题上常常更干净）。调用时机放在 `OnDocumentComplete` 里（本仓 `gohl-ui/main.go` 就是这样），签名 `gohl.SetOption(uint32(hwnd), gohl.HTMLAYOUT_FONT_SMOOTHING, 3)`。
- **平滑只在底不透明时有效**：Windows 在 layered / 每像素 alpha 的窗上会把 ClearType 降级成灰度 AA，看着更糊。gohl 的圆角是用 region 实现的（`setRoundedRegion`，`htmlayout_ui.go:1163`），`WS_EX_LAYERED` 只是定义了、全绑没有调用点，所以默认路径不踩这个坑；只有开 `HTMLAYOUT_TRANSPARENT_WINDOW`（选项号 6）或自己加 layered 才会遇到。


## Units

| Unit | Description | Example |
|------|-------------|---------|
| `px` `pt` `pc` `cm` `mm` `in` | Absolute | `width: 100px;` |
| `%` | Relative to parent | `width: 80%;` |
| `dip` | Device-independent pixel (recommended) | `height: 32dip;` |
| `%%` | Remaining space % (HTMLayout unique) | `width: 50%%;` |
| `*` / `N*` | Flex share, with optional weight | `height: *;` `width: 3*;` |
| `auto` | 等价于 `1*`（`flows/content/flow-template.htm` 原文："width:auto (that is 1*)"） | `width: auto;` |
| `min-intrinsic` `max-intrinsic` | 内容收缩/撑满宽 | `width: max-intrinsic;` |
| `calc()` | 真 CSS，纯算术部分无需脚本即可用（详见 CSSS! 一节） | `width: calc(50% + 40px);` |
| `em` `ex` | Relative to font size | `font-size: 1.2em;` |

> **`%` vs `%%` vs `*`**: `%` = parent-relative; `%%` = remaining free space %; `*` = flex share (`*` = `1*`), `3*` = 3 weight shares.
> `dip`：`csss!/dip-units.htm` 定义 `1dip == 1/96 inch`（96dpi 下 1dip=1px，120dpi 下 100dip=125px）。
> 弹簧不只用在 width/height：`margin-top:*` `padding-bottom:*` 同样合法。

> **`min-width:auto` —— 多行对不齐的根因**：`cssmap.htm` 明写 HTMLayout 给**所有元素**默认 `min-width:auto`（= min-intrinsic，"to mimic Internet Explorer behavior"），`<table>` 默认 `max-width:auto`。所以在 `flow:horizontal` 里排的一行行 div，每格都不会窄于自己的内容宽度，长路径那一行会把整列顶开，各行列宽互不相同。**要跨行对齐就用 `<table>`**（列宽由表统一求解），或者给每格写死 `min-width:0`。

## flow Layout (replaces flexbox)

| Value | Description |
|-------|-------------|
| `vertical` | 单列，**CSS 默认布局** |
| `horizontal` | 单行（不换行） |
| `h-flow` / `horizontal-flow` | 多行，即砖墙布局（**换行的是它**） |
| `v-flow` / `vertical-flow` | 多列（溢出时换到**下一列**，不是下一行） |
| `flow:"a c" "b c"` | 模板网格，子元素用 `float:"a"` 落位 |
| `flow:grid` | 位置网格，用 `top:1#; left:2#; right:3#` 指定行列 |
| `flow:table-fixed` | 固定列宽的表格流（`<table>` 上的等价物是 `fixedlayout` 属性） |

换行触发条件（`flows/content/flow-horizontal-flow.htm` 原文）：一行里 flex 之和 > `1*`（100%%）、或用了 `clear:left|right|both`、或横向空间不足。

```css
#app { size: *; flow: vertical; border-spacing: 10dip; }
.toolbar { flow: horizontal; border-spacing: 8dip; vertical-align: middle; }
li { width: 30%; height: width(30%); }               /* square */
.item:nth-child(3n) { clear: right; }                 /* grid: newline every 3 */
```

> `border-spacing` 在 flow/网格容器上就是 `gap`；官方样例一律**同时**写 `padding`（例：`padding:4px; border-spacing:4px`），说明它只算元素之间、不含外缘（外缘行为未证实）。轴向版本 `border-spacing-x` 有样例（`forms/tree-view.htm`），`border-spacing-y` 未见。负值合法（`flows/template-layout.htm` 用 `-1px`）。

> **flow 容器里的 `<span>` 不会分行**：行内元素会被收进同一个行盒，`flow:vertical` 也压不住，几行内容挤成一行。要么改用 `<div>`，要么给每个子元素 `display:block`。

> **左中右三段式（标题栏 / 工具条的靠右布局）**：要「左内容 + 中间撑开 + 右内容」，父容器里**只放三个元素**——左块、`width:*` 的 spacer、右块，**右块写 `width:auto`**。
>
> **三个都必须是自己带 `display:block` 的元素（`<div>`，或被 `<div>` 包起来的元素）**。不能直接摆 `<button>`、`<span>` 这类不是 block 的：它们会被收进同一个行盒，spacer 拿不到自己那一格，撑不到最大。右侧几个按钮因此要包进一个容器（如 `.win-opt`），不与 spacer 平铺。
> ```html
> <style>
>   #tbar{ flow:horizontal; height:34dip; padding:0 10dip; border-spacing:6dip; vertical-align:middle; }
>   #tbar .spacer{ width:*; }
>   .win-opt{ width:auto; text-align:right; }   /* 右块：block + auto 宽 + 居右 */
> </style>
> <div id="tbar" -gohl-drag>
>   <div class="t">图标字体预览</div>
>   <div class="spacer"></div>
>   <div class="win-opt">
>     <button id="win-min" class="wbtn" -gohl-min>─</button>
>     <button id="win-close" class="wbtn">✕</button>
>   </div>
> </div>
> ```
> 反过来，把 `<button>` 直接当左块或右块（不套 div）就会看到 spacer 撑不满、按钮不靠右——布局没错，是元素不是 block。
> 另一个右推法（官方样例常用）：给容器 `padding-right:100%%`（`%%` = 剩余空间的百分比，不是笔误），见 `abs/menus.htm`。

## Alignment

| Property | Values |
|----------|--------|
| `horizontal-align` | `left` / `center` / `right` |
| `vertical-align` | `top` / `middle` / `bottom` |
| `text-align` | `left` / `center` / `right` |

```css
.btn {
    flow: horizontal;
    horizontal-align: center;
    vertical-align: middle;
    text-align: center;
    behavior: clickable;
}
```

> `horizontal-align` / `vertical-align` 在 `horizontal-vertical-align/` 三个样例里都是**设在容器上**的。要让 flow 里的**某个子元素**单独靠上/靠下/居中，用弹簧而不是 `vertical-align`：`margin-top:*` → 沉底，`margin-bottom:*` → 置顶，两者都写 → 居中（`flows/content/flow-horizontal.htm` 有原话说明）。叶子元素上的 `vertical-align` 仍然有效（`generic/sidebars.htm`）。

## Backgrounds & Gradients

> **No `linear-gradient()`** (silently ignored, falls back to solid). Use four-corner colors:

```css
/* Format: top-left top-right bottom-right bottom-left */
background-color: red yellow blue yellow;

#title-bar-wrap { background-color: #4ecdc4 #44a08d #282b32 #1a1a2e; }

/* Adjacent colors without space also work */
background: #01A0FB#29F499 #01A0FB#29F499;
```

### Background Image Modes

| Mode | Description |
|------|-------------|
| `no-repeat` | No scale, no repeat |
| `repeat` | Tile |
| `repeat-x` / `repeat-y` | Single direction repeat |
| `expand` | Nine-patch (corners preserved) |
| `stretch` | Stretch to fit (add `keep-ratio`) |

```css
background: url(bg.png) expand;
background-position: 30px 30px 30px 30px;  /* top right bottom left */
background: url(image.png) stretch keep-ratio;
background-image: url(data:image/png;base64,iVBORw0KGgo...);
```

### Foreground

Same syntax as background, renders above content. Used for button icons.

```css
foreground: red url(fg.png) no-repeat 50% 50%;
/* Layer order: parent foreground > child background > parent background */
```

## Image Transformations

```css
background-image-transformation: opacity(0.5);
background-image-transformation: colorize(#FFCC00);
background-image-transformation: flip-x();
background-image-transformation: flip-y();
background-image-transformation: contrast-brightness-gamma(0.5, 0.3, 1.2);

/* Icon state coloring */
.button.danger   { foreground-image-transformation: colorize(#f87171); }
.button:disabled { foreground-image-transformation: colorize(#6b7280); }
```

## Borders & Glow

```css
border-radius: 8dip;
border-radius: 50%;                    /* circle */
border-radius: 8dip 8dip 0 0;         /* top round, bottom square */

outline: 2dip glow #6366f1 2dip;      /* glow effect */
#input:hover { outline: 2dip glow #94bdc7 2dip; }
```

## Text

```css
font: system;           /* system input font */
font: system-menu;      /* system menu font */
text-selection: #FFF #C00;   /* foreground background */
text-transform: uppercase;
content: "hover";       /* replace element text (CSS-only) */
```

> **截断省略号要三条一起写**，且是 `overflow-x:hidden` 而不是 `overflow:hidden`——`behaviors/path-behavior.htm` 里原作者留了原话注释 `// these three must be used together`：
> ```css
> .path { overflow-x: hidden; white-space: nowrap; text-overflow: ellipsis; }
> ```
> 同样写法见 `grid/sortable-grid.htm`、`forms/listview.htm`。

> **Not implemented:** `letter-spacing`, `text-decoration: blink`, `direction`

## behavior (Core Extension)

> **原生表单控件不用手写 behavior**：`<button>`、`<input type=button|text|password|number|checkbox|radio>`、`<select>`、`<textarea>` 在 `forms/form.htm`、`forms/basiccontrols.htm`、`forms/buttons.htm` 里都是裸标签直接用。只有把**非表单元素**（`<div>`/`<td>`/`<span>`/`<p>`）当控件时才要显式声明。注意是 `behavior:edit`，**没有** `behavior:editable`；数字框是 `behavior:number`。

| behavior | Purpose |
|----------|---------|
| `clickable` | Clickable (no focus), **required for OnButtonClick** |
| `button` | Standard button（也能给 `<td>`/`<span>` 用，`forms/buttons.htm:58` 就把表格单元格声明成按钮） |
| `check` / `radio` | Checkbox / radio |
| `number` | 数字输入（`csss!/total.htm`） |
| `edit` | Text input, supports `filter`, `novalue` |
| `select` | Dropdown |
| `menu` / `menu-bar` / `popup-menu` | Menus |
| `grid` | 表格行为：排序、选中行、`fixedrows` 表头冻结（见下面表格一节） |
| `path` | 文件路径输入 + 浏览按钮（`behaviors/path-behavior.htm`） |
| `progress` | Progress bar |
| `light-box-dialog` | Modal dialog (Enter=ok, Esc=cancel) |
| `password` | Password field (`password-char`) |
| `hyperlink` | Hyperlink (triggers OnHyperlinkClick) |
| `tabs` | Tab switching |
| `scroller` | 拖拽滚动（`generic/overflows.htm`） |

```html
<input filter="0~9" novalue="Only digits" />
<!-- filter whitelist; ^ prefix = exclude, e.g. "^.,-" -->
<style>input:empty { color: #555; }</style>  /* placeholder style */
```

> **没有 `placeholder` / `validation` 属性**（全树 0 命中），等价物是 `novalue="…"` + `:empty`。其他已证实的表单属性：`maxlength`、`password-char="#"`、`type=number step/minvalue/maxvalue`、`type=decimal format="grouping:3; fdigits:4"`、`[negative]`/`[invalid]` 状态属性、`readonly`、`accesskey`、`checked`/`mixed`/`checked=undefined`（三态）。`<label>` 标签在样例里根本不出现（标题用裸文本或 `<caption>`），`for=` 是否生效未证实。

### gohl 侧的事件钩子

样例里的 `click!:` / `value-changed!:` / `$()` 都属于 CSSS!+TIScript，**无脚本宿主不可用**。gohl 走的是 HTMLayout 原生通知码（`D:\codes\go\Gohl\htmlayout_ui.go:793-853`）：

| Go 钩子 | 原生 cmd | 触发元素 |
|---------|----------|----------|
| `OnButtonClick` | `BUTTON_CLICK` | `<button>` / `behavior:button|clickable` |
| `OnButtonStateChanged` | `BUTTON_STATE_CHANGED` | checkbox / radio |
| `OnSelectionChanged` | `SELECT_SELECTION_CHANGED` | `<select>` / `<option>` |
| `OnValueChange` | `EDIT_VALUE_CHANGED` | `<input>` / `<textarea>` |
| `OnVisibleChange` | `VISIUAL_STATUS_CHANGED` | — |

所以「按 `id` 前缀在 Go 里路由点击」是正确做法，不是绕路。

## @set Custom Scrollbar

```css
@set small-v-scrollbar {
    :root { background-color: #131417; }              /* 整条滚动条的底板 */
    .base { width: 4dip; }                            /* 竖直滚动条用 width；水平用 height */
    .slider { background-color: #2a2e42; min-height: 32dip; }
    .slider:hover { background-color: #6366f1; }
    .prev, .next { background-color: transparent; height: 0; border: none; }
    .prev:disabled, .slider:disabled, .next:disabled { height: 0; border: none; foreground-image: none; }
}

pre#log { overflow: auto; vertical-scrollbar: small-v-scrollbar; }
```

部件类（`css-plus/scrollbar-styling.htm` 全集）：`:root`、`.base`(轨道)、`.slider`(滑块)、`.prev`(上/左)、`.next`(下/右)、`.prev-page`、`.next-page`、`.corner`(横竖交汇角)。状态伪类 `:hover` `:active` `:disabled` 都可用。

> **值不带 `@` 前缀**：写 `vertical-scrollbar: small-v-scrollbar`。全树样例无一处用 `@name`，`vertical-scrollbar:@name` 这种写法未获证实。可全局套用：`* { vertical-scrollbar: small-v-scrollbar; }`（`small-scrollbar-styling.htm`）。
> `@set` 可继承复用：`@set popup-menu < std-popup-menu { … }`（`forms/edits.htm`）。
> `foreground-image: set(name)` 这种引用方式在样例里没找到，未证实。

## Selectors

**Attribute:** `[foo^="val"]` prefix, `[foo$="val"]` suffix, `[foo*="val"]` contains

**Pseudo-classes:**
- `:link` / `:visited` (mutually exclusive, elements with href)
- `:active` / `:hover` (any class)
- `:focus` (elements with tabindex)
- `:not(...)` `:nth-child(An+B)` `:nth-last-child(An+B)`
- `:checked` `:disabled` `:read-only` `:empty`
- `:only-child` `:only-of-type` `:has-child` `:has-children`
- `:current` (selected option)
- `:tab-focus` (focused via TAB key)

**Drag & drop:** `:popup` `:owns-popup` `:drop-target` `:drag-over` `:moving` `:copying` `:drag-source` `:drop-marker`

## transition Animation (Syntax differs from CSS3)

Naturally **bidirectional** — forward on style gain, auto-rollback on style loss.

```css
/* Basic: property(easing, duration[, delay]) */
div { transition: width(linear, 0.4s) height(linear, 0.4s); }

/* Named params (order-independent) */
div { transition: width(easing-function: linear, duration: 0.4s, delay: 100ms); }

/* Separate in/out */
div {
    transition: width(easing-function-in: linear,  duration-in: 0.4s,
                      easing-function-out: cubic-in-out, duration-out: 200ms);
}

/* Forward only (suppress rollback) */
div { transition: size(linear, 500ms, none); }

/* Rollback only */
div { transition: size(none, linear, 500ms); }

/* Image crossfade */
img:hover { foreground-image: url(red.png); transition: blend; }

/* Transition sound */
div { transition: sound(in: url(a.wav), out: url(b.wav)); }
```

**Easing functions:** `linear` `quad` `cubic` `quart` `quint` `sine` `expo` `circ` `elastic` `back` `bounce`, each with `-in` / `-out` / `-in-out`.

**Animatable properties:** `width` `height` `size` `background-color` `opacity` `margin` `border` `foreground-image` (with blend), flex values (`3*`).

## Constants & Colors

```css
@const BASE_COLOR: #22c55e;   /* Cannot be overridden; can reference each other */
@const MY_BORDER: 3px dashed @BASE_COLOR;

color: tint(@BASE_COLOR, 0.25);       /* tint(color, brightness, saturation[optional]) */
color: hsl(30, 100%, 50%);
background-color: rgba(255,0,0,0.5);
div { opacity: 0.25; }

/* Multi-color = gradient transparency */
background-color: rgb(255,0,0) rgba(255,0,0,0.15) rgba(255,0,0,0.15) rgb(255,0,0);
```

## @image-map (DPI-aware + Sprite)

```css
@image-map dpi-aware {
    src: url(star-1x.png) 100dpi, url(star-2x.png);
}
#star { size: 180dip; background-image: image-map(dpi-aware); }

@image-map tb-icon {
    src: url(rttb.png);
    cells: 15 2;                        /* 15 cols, 2 rows */
    items: bold, italic, underline;     /* row-major */
}
toolbar > button.bold { foreground-image: image-map(tb-icon, bold); }
```

## CSSS! Scripting

> **gohl 用不了本节。** 绑定层没有任何脚本入口（`D:\codes\go\Gohl` 里 `Invoke` / `CallFunction` / `Eval` / `ScriptValue` 全为 0 命中），所以 `xxx!` 属性、`$()` 选择器、`@()` 匿名函数在 gohl 宿主里**无法验证也无法从 Go 侧配合**。把本节当引擎参考读，别把功能设计在它上面——事件一律走 Go 的 `OnButtonClick` 等钩子。
> 真正对 gohl 有用的是下面的 `calc()`：它是**纯 CSS**，不需要脚本。

Event properties (全树 grep 到的完整族): `assigned!` `click!` `double-click!` `hover-on!` `hover-off!` `focus-on!` `focus-off!` `active-on!` `active-off!` `value-changed!` `animation-start!` `animation-end!` `animation-step!`

```css
a {
    hover-on!:  $(p.slave):hover = true;    /* csss!/1.htm：一条 ! 语句去改别的元素 */
    hover-off!: $(p.slave):hover = false;
}
```

```css
.notification.show {
    display: block;
    assigned!: self::opacity = 0.01, self.start-animation();
    animation-step!: self::opacity < 1.0
        ? ( self::opacity = self::opacity + 0.01, return 3 )
        # ( self.fade = "", return cancel );
}
```

> `self::prop` accesses property; ternary uses `? (a) # (b)` (`#` not `:`); `,` separates statements

### calc()

`csss!/calc-basic.htm`、`calc-ext.htm`、`calc-dom.htm`. 三种层次，可用性不同：

```css
/* 1) 纯 CSS 算术 —— gohl 可直接用 */
@const WIDTH_EXPR: calc(50% + 40px);   /* 可存进常量再引用（calc-ext.htm:4,16） */
width: calc(50% + 40px);
width: @WIDTH_EXPR;
height: calc(2px + 25%);

/* 2) 引擎内建测量函数 —— 布局期求值 */
width:  calc(max-intrinsic-width() + system-scrollbar-width());
height: calc(intrinsic-height() + 40px);

/* 3) DOM 相对表达式 */
width: calc(self:index * 10%);         /* 第 n 个子元素 → n*10% */
width: calc(200px + self:value * 10px);
```

> 第 3 类里 `self:value` 这类**随输入变化**的表达式，样例是靠 `value-changed!: self.update();` 强制重算的（`calc-ext.htm:21`）——那一步需要脚本，gohl 做不到。要用就只取第 1、2 类，或在 Go 侧改样式。

## position

Supports `static`, `relative`, `absolute`, `fixed` — same as CSS 2.1.

- `position:fixed` 的包含块是**整个视图**，滚动不跟着走 → 全屏遮罩/面板的标准写法（`behaviors/light-box-dialog.htm`）：
  ```css
  div.shim { position: fixed; top: 0; right: 0; bottom: 0; left: 0; background-color: rgba(0,0,0,0.5); }
  ```
- `position:absolute` 相对最近的 `position:relative` 祖先，并随其内容滚动（`abs/rel_abs.htm`、`abs/abs.htm`：嵌在普通流里的绝对元素会滚走，挂在 body 上的才固定）。下拉菜单就是 `li{position:relative}` + 子 `menu{position:absolute;display:none}`。
- 居中不用 flex，用四向弹簧：`left:*; right:*; width:100px; top:*; bottom:*;`（`abs/fix_center.htm`）。
- **`z-index` 在整个样例树里 0 命中，未获证实。** 层叠顺序按文档顺序 + 定位元素后画来保证；需要盖住别人就把它放到 DOM 末尾，或用 `position:fixed` 的遮罩。别依赖 `z-index` 调层级。

## SVG

**No inline `<svg>`** in HTML. `.svg` files used as image resources only (rasterized).

## 表格：跨行对齐的列只能用 `<table>`

一行行 `<div>` + `flow:horizontal` **做不到**跨行列对齐（每格 `min-width:auto` 不会被压到内容宽度以下，见 Units 一节）。官方样例里也没有 `display:table` 给 div 用的写法（全树 0 命中）。要真表格就上 `<table>`。

### 支持的标记

`<table>/<thead>/<tbody>/<tfoot>/<tr>/<th>/<td>`、`colspan`、`rowspan`、`<caption>`（放在它所修饰的 `<tbody>` 前面）、多个并列 `<tbody>`（`<tbody .first>` 这种 id/class 简写）。结束标签可省略（`grid/scrollable-table.htm` 就这么写）。
**没有** `<col>`/`<colgroup>`，**没有** `table-layout` 属性——固定列宽的对应写法是 `flow:table-fixed` 或表格上的 `fixedlayout` 属性。

### 列宽

宽度写在**表头单元格**上（有 `<thead>` 时），没有表头就写 `td:nth-child(n)`：

```css
table { flow: table-fixed; }              /* 等价于 <table fixedlayout> */
td:nth-child(1) { width: 20%; min-width: 4em; }
td:nth-child(2) { width: 200px; }
td:nth-child(3) { width: *; }             /* 吃掉剩余全部宽度 */
/* 加权分配：width:1* / width:2*（grid/sortable-grid.htm）；也可用 10%%/25%%/50%% */
```

`dip` 在任何表格里都没被用在单元格宽度上（只用过 `px`/`em`/`%`/`*`/`N*`/`%%`）。

### 表头冻结 + 滚动（三种已证实的做法）

```css
/* A) 原生网格行为：最常用，还白送排序和选中行 */
table { behavior: grid; overflow: auto; width: 100%%; height: 100%%; }   /* + <table fixedrows="1"> */

/* B) thead/tbody 分工，只让 tbody 滚 */
table { flow: table-fixed; }
table > tbody { height: *; overflow-y: auto; }

/* C) 整表滚动，不要表头 */
table#t { width: *; height: *; overflow: auto; }
```

没有 `position:sticky` 这类机制。

### 边框、斑马纹、悬停

```css
table { border-spacing: 0; }                     /* 网格里几乎都这么写；也可 <table cellspacing="0"> */
tr:nth-child(odd) { background-color: #F4F3F9; }  /* 斑马纹，fixed-table/table-fixed-10000-rows.htm */
tr:hover { background-color: white white orange orange; }   /* 4 值=四角色，csss!/table-checkboxes.htm */
td:hover { background-color: yellow; }            /* 多数网格样例悬停的是 td */
tr:current { background-color: highlight; }       /* 网格选中行用 :current，不是 :checked */
```

`tr:hover` 设背景**确实生效**，但官方网格样例更常悬停 `td`；保险做法是两边都写或只写 `td`。`border-collapse` 只在 `cssmap.htm` 的属性表里出现（注为 "as per CSS"），无样例演示。

### 大行数

`stress/` 里万行表格是 `.jsp` 预生成的静态标记（单文件 1.1MB），**没有虚拟化**，靠 `behavior:grid` + `overflow:auto` + `fixedlayout` 硬扛。唯一真正的懒加载是 `behavior:virtual-grid`（`grid/virtual-table.htm`），但它要 TIScript 数据源，**gohl 无脚本环境用不了**。所以行数上限要自己在 Go 侧截，并在界面上写明"只画前 N 行，计数已含全部"。

## 提示气泡 tooltip（无需脚本）

```html
<div title="纯文本气泡">…</div>
<div tooltip="可以放 <b>标记</b>">…</div>
<div titleid="my-tip">…</div>
<popup id="my-tip" style="display:none"><p>HTML 气泡内容</p></popup>
```

```css
_service > popup[role=tooltip] { background: #20232a; color: #e9ebee; border: 1dip solid #e8c04a; }
```

`title` 直接可用（`tooltips/titles.htm`），样式挂在 `_service > popup[role=tooltip]` 上；气泡元素自身带 `:popup` 伪类。没有 `behavior:tooltip`。

## Tables & Lists

```css
ul { list-style: disc outside url(bullet.png); }
ol { list-style: decimal; }
```

## mapping — RTL Auto-mirror

```css
.tool-button { mapping: left-to-right(margin, foreground-image); }
```
