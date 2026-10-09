# HTMLayout 官方样例库索引

根目录：`D:\software\aardio\example\WebUI\HTMLayout\html_samples`

这是 aardio 随包带的 HTMLayout 官方样例集（约 40 个主题目录）。**要确认某个 CSS/behavior 到底支不支持，先来查这里，别拿浏览器 CSS 的经验猜。**

## 两个必须知道的坑

- **`HTML32.htm` 不是样例**：它是 W3C HTML 3.2 规范正文，被拿来当渲染压力测试，里面**没有** `flow:` / `size:` / `behavior:`。想查属性支持情况，看 **`cssmap.htm`**（HTMLayout 专有的行用 `#ffe4ca` 底色标出）。
- 样例文件可能是 GBK 编码，中文注释会花屏——读代码不读注释即可，别在编码上耗时间。

## 目录 → 能证明什么

| 目录 | 内容 |
|------|------|
| `flows/` | flow 的全部取值：`horizontal` / `vertical` / `horizontal-flow`（换行=砖墙）/ `vertical-flow`（换列）/ 模板 `flow:"a c" "b c"` / `flow:grid`（`top:1#;left:2#`）。`flows/content/*.htm` 每个带官方英文说明 |
| `horizontal-vertical-align/` | 容器级 `horizontal-align` / `vertical-align`；flow 子元素靠弹簧对齐，不是 `vertical-align` |
| `abs/` | `position:absolute/relative/fixed`；`fix_center.htm` 用 `left:*;right:*` 居中；`menus.htm` 用 `padding-right:100%%` 把内容顶到左边 |
| `fixed-table/` | 表格：`flow:table-fixed`、`rowspan`、多 `tbody`（`<tbody .first>`）、`caption` 分组、10000 行带滚动 |
| `grid/` | `behavior:grid` + `fixedrows="1"` 的表头冻结网格、`sortable-grid`、`virtual-table`（**要 TIScript，无脚本环境不可用**） |
| `stress/` | 万行表格 / 万项 `<select>` / 16 层嵌套表。行数由 `.jsp` 预生成，**没有虚拟化**，靠 `behavior:grid`+`overflow:auto`+`fixedlayout` |
| `css-plus/` | `@const`、`@set` 滚动条样式（`:root` + 各部件类）、`theme:` URL、`@import`、`@media screen && composition-supported` |
| `csss!/` | CSSS! 迷你语言：`calc()`、`$()` 选择器表达式、`event!:` 声明式事件、`dip-units.htm`（`1dip == 1/96 inch`） |
| `@font-face/` | 只有一个 `test.htm`：`@font-face{font-family:NinjaLine;src:url(NinjaLine.ttf)}` + `font:18pt NinjaLine` |
| `forms/` | 原生控件（`button/input/select/textarea`）**不写 behavior 也能用**；`select > caption`/`> popup`/`> option` 样式、`multiple="checks"`、`novalue`、`filter`、`maxlength`、三态复选 |
| `behaviors/` | `behavior:clickable/button/check/edit/number/path`、`light-box-dialog`（`div.shim{position:fixed}` 遮罩）、`treeview`、`dropdown`、`actions.htm`（提到原生 `BUTTON_CLICK`） |
| `tooltips/` | `title="…"` 纯声明式气泡（无需脚本），`tooltip="…"` 可放标记，`titleid="popup-id"` 绑隐藏 `<popup>`；样式挂 `_service > popup[role=tooltip]` |
| `menu/` | `<button type="menu">` + 嵌套 `<menu>` 的无脚本弹出菜单；`:owns-popup` |
| `transitions/` | `transition: width(linear,400ms)`、`transition:blend`、`transition:slide`；**`@keyframes` 全树 0 命中** |
| `fieldset/` `effects/` `editor/` `richtext/` `includes/` `for-at/` `goodies/` `optimizations/` `printing/` `svg/` `animated-png/` `drag-n-drop/` `sort-of-frames/` `frames/` `communication/` `back/` `border-radius/` `animations/` `generic/` `behaviors/` | 其余专题：字段组、跑马灯、富文本、`<include src>`、`-for/-at` 控件配对、`data:` URL、SVG 只能当图片资源 |

## 全树 grep 结果为 0 的东西（别写进代码）

`table-layout`（HTMLayout 用 `flow:table-fixed`）、`<col>` / `<colgroup>`、`display:table`、`position:sticky`、`z-index`、`@keyframes`、`@renderer`、`@for` / `@each`、`@msg` / `@prelude`、`placeholder` / `validation` 属性（等价物是 `novalue` + `:empty`）、`<label>` 标签（样例里根本不出现，标题用裸文本或 `<caption>`）、`resources://` / `embed://`（这两个是 gohl 绑定层自己加的 scheme，不是 HTMLayout 标准）。

## 无脚本（gohl）环境下可用 / 不可用

- **可用**：全部 CSS 布局与状态伪类、`title` 气泡、`transition`、`<button type=menu>`+`<menu>`、`behavior:grid` 的排序与表头冻结、`@const`/`@set`/`@import`/`calc()`、`<widget type=vscrollbar for=… at=…>`。
- **不可用**：`behavior:virtual-grid`（数据源要 TIScript）、一切 `event!:`（`click!:` / `value-changed!:` …）和 `$()` 表达式、`<include src>`（无 base URL 时能否解析未证实）。
