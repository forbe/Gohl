package gohl

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"
	"testing"
	"unsafe"
)

// 这批测试盯的是「iconfont 在 HTMLayout 里到底能不能用」这件事：PUA 别名补得对不对、
// cmap 重造后读不读得回来、家族名改没改、内置那份 zip 里的字体是不是可用状态。
// 对照素材是两支真实的 iconfont：
//   testdata/iconfont.old.ttf —— 手工开好 56xx 别名的老字体（168 个码位）
//   testdata/iconfont.new.ttf —— iconfont.cn 直接下载、整支挂在 PUA 的新字体

const (
	newFontPath = "testdata/iconfont.new.ttf"
	oldFontPath = "testdata/iconfont.old.ttf"
)

func readFont(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustGlyphs(t *testing.T, data []byte) []Glyph {
	t.Helper()
	list, err := GlyphsOf(data)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

// 新下载的字体整支挂在 PUA，HTMLayout 画出来是 '?'。AliasIconfont 给它补一份
// 56xx-59xx 别名；这一测盯两件事：每个 PUA 字形是不是都拿到了别名，以及别名是不是真
// 指向同一个字形。
func TestAliasNewFont(t *testing.T) {
	raw := readFont(t, newFontPath)
	before := mustGlyphs(t, raw)

	pua := []Glyph{}
	uniqGid := map[uint16]bool{}
	for _, g := range before {
		if g.Codepoint >= puaLo && g.Codepoint <= puaHi {
			pua = append(pua, g)
			uniqGid[g.GlyphID] = true
		}
	}
	if len(pua) == 0 {
		t.Fatal("这份字体里没有 PUA 码位，测试前提变了")
	}

	patched, n, err := AliasIconfont(raw)
	if err != nil {
		t.Fatalf("AliasIconfont: %v", err)
	}
	after := mustGlyphs(t, patched)

	byCp := make(map[rune]uint16, len(after))
	for _, g := range after {
		byCp[g.Codepoint] = g.GlyphID
	}
	if n != len(uniqGid) {
		t.Fatalf("造了 %d 个别名，PUA 里不同字形 %d 个，没造全", n, len(uniqGid))
	}
	if len(after) != len(before)+n {
		t.Fatalf("重写后 %d 个码位，原来是 %d 个 + 别名 %d 个", len(after), len(before), n)
	}

	// 按老公式该落在哪的，点名核对；其余的只要求同 gid 的别名存在于 56xx-59xx。
	gidAlias := map[uint16][]rune{}
	for r, gid := range byCp {
		if r >= aliasLo && r <= aliasHi {
			gidAlias[gid] = append(gidAlias[gid], r)
		}
	}
	for _, g := range pua {
		if g.Codepoint >= puaAliasLo && g.Codepoint <= puaAliasHi {
			want := g.Codepoint - aliasOff
			if byCp[want] != g.GlyphID {
				t.Fatalf("U+%04X 的公式别名 U+%04X 没挂上（那里 gid=%d）", g.Codepoint, want, byCp[want])
			}
			continue
		}
		if len(gidAlias[g.GlyphID]) == 0 {
			t.Fatalf("U+%04X (gid=%d) 一个 56xx 别名都没拿到", g.Codepoint, g.GlyphID)
		}
	}
	// 原码位必须一个不少：别名是加上去的，不是搬走的。
	for _, g := range pua {
		if byCp[g.Codepoint] != g.GlyphID {
			t.Fatalf("U+%04X 重写后变了（gid=%d，期望 %d）", g.Codepoint, byCp[g.Codepoint], g.GlyphID)
		}
	}
	t.Logf("PUA %d 个 → 别名 %d 个，重写后共 %d 个码位", len(pua), n, len(after))
}

// 内置那份本来就该是补好别名的状态：再处理一次不该动它一根手指。
func TestBundledNeedsNoAlias(t *testing.T) {
	data, err := ReadResource("iconfont.new.ttf")
	if err != nil {
		t.Fatal(err)
	}
	patched, n, err := AliasIconfont(data)
	if err != nil {
		t.Fatalf("内置字体不该报错: %v", err)
	}
	if n != 0 {
		t.Fatalf("内置字体已有 56xx 别名，却又造了 %d 个", n)
	}
	if !bytes.Equal(patched, data) {
		t.Fatalf("没造别名却改了字节: %d → %d", len(data), len(patched))
	}
}

// 内置那份必须能在页面上直接用：家族名是 iconfont，56xx 段有货，PUA 原件也还在。
func TestBundledIconfontUsable(t *testing.T) {
	f, err := BundledIconfont()
	if err != nil {
		t.Fatal(err)
	}
	if f.Aliases != 0 {
		t.Errorf("内置字体应该在打包时就补好别名，这里又造了 %d 个", f.Aliases)
	}
	if f.Family == "" {
		t.Error("家族名是空的")
	}
	orig, err := FontFamily(readFont(t, newFontPath))
	if err != nil {
		t.Fatal(err)
	}
	if f.Family == orig {
		t.Errorf("BundledIconfont 该换成唯一名，免得和页面里别的 face 撞车，结果还是 %q", f.Family)
	}
	usable := 0
	for i, g := range f.Glyphs {
		if i > 0 && !(g.Codepoint > f.Glyphs[i-1].Codepoint) {
			t.Fatalf("第 %d 条码位没按升序排：%v 跟 %v 后面", i, g, f.Glyphs[i-1])
		}
		if g.Codepoint >= aliasLo && g.Codepoint <= aliasHi {
			usable++
		}
	}
	if usable < 300 {
		t.Errorf("56xx 段只有 %d 个码位，内置字体的别名没打好", usable)
	}
	if css := f.CSS(f.Serve("gohl-test")); !bytes.Contains([]byte(css), []byte(f.Family)) {
		t.Errorf("CSS 里没带上家族名: %s", css)
	}
}

// 单独盯 cmap 编码器：拿老字体自己的码位表重造一张 cmap 再读回来，必须一字不差。
// 别名逻辑对不对是另一回事，这条保证 format 4 编码 + 表目录改写没写坏。
func TestCmapRebuildRoundTrip(t *testing.T) {
	font := readFont(t, oldFontPath)
	orig := mustGlyphs(t, font)

	byCp := make(map[rune]uint16, len(orig))
	for _, g := range orig {
		byCp[g.Codepoint] = g.GlyphID
	}
	cmap, err := buildCmap(byCp)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := replaceCmap(font, cmap)
	if err != nil {
		t.Fatal(err)
	}
	after := mustGlyphs(t, rebuilt)

	if len(after) != len(orig) {
		t.Fatalf("码位数 %d，原来是 %d", len(after), len(orig))
	}
	got := make(map[rune]uint16, len(after))
	for _, g := range after {
		got[g.Codepoint] = g.GlyphID
	}
	for _, g := range orig {
		if got[g.Codepoint] != g.GlyphID {
			t.Fatalf("U+%04X 的 gid 从 %d 变成了 %d", g.Codepoint, g.GlyphID, got[g.Codepoint])
		}
	}
}

// 老字体是手工别名的样板：码位数和几个标志性名字必须还认得，不然说明读表逻辑坏了。
func TestOldFontCmap(t *testing.T) {
	list := mustGlyphs(t, readFont(t, oldFontPath))
	if len(list) != 168 {
		t.Fatalf("码位数 %d，期望 168", len(list))
	}
	for _, want := range []struct {
		cp   rune
		name string
	}{{0x5666, "arrow-right-circle"}, {0x561D, "jubao"}} {
		found := false
		for _, g := range list {
			if g.Codepoint == want.cp {
				if g.Name != want.name {
					t.Fatalf("U+%04X 名字 %q，期望 %q", want.cp, g.Name, want.name)
				}
				found = true
			}
		}
		if !found {
			t.Fatalf("U+%04X 不在 cmap 里", want.cp)
		}
	}
}

func TestRenameFontChangesFamily(t *testing.T) {
	raw := readFont(t, newFontPath)
	old, err := FontFamily(raw)
	if err != nil {
		t.Fatal(err)
	}
	out, err := renameFont(raw, "gohl-icon-7")
	if err != nil {
		t.Fatal(err)
	}
	got, err := FontFamily(out)
	if err != nil {
		t.Fatal(err)
	}
	if got != "gohl-icon-7" {
		t.Errorf("改名后自报 %q，想要 gohl-icon-7（原来叫 %q）", got, old)
	}
	if old != "iconfont" {
		t.Errorf("示例字体家族名变了：%q", old)
	}
	// 改名只动 name 表：码位表必须一张不差。
	a, b := mustGlyphs(t, raw), mustGlyphs(t, out)
	if len(a) != len(b) {
		t.Fatalf("改名后码位数 %d，改名前 %d", len(b), len(a))
	}
	for i := range a {
		if a[i].Codepoint != b[i].Codepoint || a[i].GlyphID != b[i].GlyphID {
			t.Fatalf("第 %d 条码位变了: %v -> %v", i, a[i], b[i])
		}
	}
}

// 表目录改偏移这件事必须真的改在目录项上，不能只追加字节。
func TestReplaceTableUpdatesDirectory(t *testing.T) {
	raw := readFont(t, newFontPath)
	out, err := renameFont(raw, "t-1")
	if err != nil {
		t.Fatal(err)
	}
	base := sfntBase(raw)
	o1, _, _, err := tableAt(raw, base, "name")
	if err != nil {
		t.Fatal(err)
	}
	o2, l2, _, err := tableAt(out, base, "name")
	if err != nil {
		t.Fatal(err)
	}
	if o2 == o1 {
		t.Error("name 表偏移没变，说明目录项没改")
	}
	if o2+l2 != len(out) {
		t.Errorf("新 name 表 [%d,%d) 没有正好落在文件末尾 %d", o2, o2+l2, len(out))
	}
	if binary.BigEndian.Uint32(out[:4]) != 0x00010000 {
		t.Error("改完表目录后 sfnt 版本标记都不对了")
	}
}

func TestReadResource(t *testing.T) {
	data, err := ReadResource("alibaba_puhui.ttf")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100000 {
		t.Errorf("只读出 %d 字节，普惠体不该这么小", len(data))
	}
	if _, err := ReadResource("没有这个文件.xyz"); err == nil {
		t.Error("读不存在的条目应该报错")
	}
}

// 拿一份具体字体来问它到底长什么样：ICONFILE=路径 go test -run TestProbe
func TestProbe(t *testing.T) {
	path := os.Getenv("ICONFILE")
	if path == "" {
		t.Skip("没给 ICONFILE")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %d 字节，版本标记 % X", path, len(data), data[:4])
	list, err := GlyphsOf(data)
	if err != nil {
		t.Fatalf("GlyphsOf: %v", err)
	}
	buckets := map[string]int{}
	named := 0
	for _, g := range list {
		switch {
		case g.Codepoint >= 0xE000 && g.Codepoint <= 0xF8FF:
			buckets["PUA(E000-F8FF)"]++
		case g.Codepoint >= aliasLo && g.Codepoint <= aliasHi:
			buckets["CJK 别名(56xx-59xx)"]++
		case g.Codepoint < 0x2100:
			buckets["BMP 常规"]++
		default:
			buckets["其它"]++
		}
		if g.Name != "" {
			named++
		}
	}
	for k, v := range buckets {
		t.Logf("  %-22s %d", k, v)
	}
	t.Logf("  共 %d 个码位，其中带 post 名字的 %d 个", len(list), named)
	for i, g := range list {
		if i >= 12 {
			break
		}
		t.Logf("  U+%04X gid=%d name=%q", g.Codepoint, g.GlyphID, g.Name)
	}
}

// 系统字体走的是另外几条路：msyh.ttc 是 ttc 集合、simsun 只有 format 4、
// seguiemj 是带 PUA 的。工具要能照读不误。
func TestSystemFonts(t *testing.T) {
	for _, path := range []string{
		`C:\Windows\Fonts\msyh.ttc`,
		`C:\Windows\Fonts\simsun.ttc`,
		`C:\Windows\Fonts\arial.ttf`,
		`C:\Windows\Fonts\seguimj.ttf`,
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("跳过 %s: %v", path, err)
			continue
		}
		list, err := GlyphsOf(data)
		if err != nil {
			t.Errorf("%s 读不出码位: %v", path, err)
			continue
		}
		if len(list) == 0 {
			t.Errorf("%s 读到 0 个码位", path)
			continue
		}
		t.Logf("%s: %d 个码位，首个 U+%04X，末个 U+%04X", path, len(list), list[0].Codepoint, list[len(list)-1].Codepoint)
	}
}

var (
	procAddFontMem       = gdi32.NewProc("AddFontMemResourceEx")
	procRemoveFontMem    = gdi32.NewProc("RemoveFontMemResourceEx")
	procCreateFontIndirc = gdi32.NewProc("CreateFontIndirectW")
	procGetGlyphIndices  = gdi32.NewProc("GetGlyphIndicesW")
)

type logFontW struct {
	lfHeight, lfWidth, lfEscapement, lfOrientation, lfWeight     int32
	lfItalic, lfUnderline, lfStrikeOut, lfCharSet                byte
	lfOutPrecision, lfClipPrecision, lfQuality, lfPitchAndFamily byte
	lfFaceName                                                   [32]uint16
}

// gidsFor 让 GDI 按家族名挑一张 face，回答这批码位的字形号 —— 引擎看到的就是这个。
func gidsFor(t *testing.T, face string, cps []rune) []uint16 {
	t.Helper()
	lf := logFontW{lfHeight: -32}
	copy(lf.lfFaceName[:], syscall.StringToUTF16(face))
	hf, _, _ := procCreateFontIndirc.Call(uintptr(unsafe.Pointer(&lf)))
	if hf == 0 {
		t.Fatalf("按家族名 %q 建字体失败（说明这个名字没注册上）", face)
	}
	defer procDeleteObject.Call(hf)
	hdc, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, hdc)
	old, _, _ := procSelectObject.Call(hdc, hf)
	defer procSelectObject.Call(hdc, old)
	out := make([]uint16, len(cps))
	for i, r := range cps {
		u := []uint16{uint16(r)}
		procGetGlyphIndices.Call(hdc, uintptr(unsafe.Pointer(&u[0])), 1, uintptr(unsafe.Pointer(&out[i])), 0)
	}
	return out
}

// 这一条是改名存在的理由：同一个会话里先后交付两支同名家族名的字体时，GDI 一直回第一份
// 的字形号（表现就是"一部分图标对、其余显示成本字"）。换成唯一名之后必须回新那份的。
func TestRenameFontBeatsSameNameCollision(t *testing.T) {
	first := readFont(t, oldFontPath) // 自报 iconfont
	second, err := renameFont(mustAlias(t, readFont(t, newFontPath)), "gohl-icon-9")
	if err != nil {
		t.Fatal(err)
	}

	firstMap := map[rune]uint16{}
	for _, g := range mustGlyphs(t, first) {
		firstMap[g.Codepoint] = g.GlyphID
	}
	secondMap := map[rune]uint16{}
	for _, g := range mustGlyphs(t, second) {
		secondMap[g.Codepoint] = g.GlyphID
	}

	// 挑两份 cmap 都认得、但期望字形号不同的码位，这样才能分辨 GDI 用的是哪一份。
	var cps []rune
	for r, gid := range secondMap {
		if o, ok := firstMap[r]; ok && o != 0 && o != gid {
			cps = append(cps, r)
		}
		if len(cps) >= 6 {
			break
		}
	}
	if len(cps) == 0 {
		t.Fatal("找不到两份字体码位相同、字形号不同的对照点，这份测试说明不了问题")
	}

	var n1, n2 uintptr
	h1, _, e := procAddFontMem.Call(uintptr(unsafe.Pointer(&first[0])), uintptr(len(first)), 0, uintptr(unsafe.Pointer(&n1)))
	if h1 == 0 {
		t.Fatalf("注册第一份失败: %v", e)
	}
	defer procRemoveFontMem.Call(h1)
	h2, _, e := procAddFontMem.Call(uintptr(unsafe.Pointer(&second[0])), uintptr(len(second)), 0, uintptr(unsafe.Pointer(&n2)))
	if h2 == 0 {
		t.Fatalf("注册改名后的第二份失败: %v", e)
	}
	defer procRemoveFontMem.Call(h2)

	fam, err := FontFamily(second)
	if err != nil {
		t.Fatal(err)
	}
	got := gidsFor(t, fam, cps)
	for i, r := range cps {
		if got[i] != secondMap[r] {
			t.Errorf("U+%04X 拿到 gid=%d，第二份期望 %d、第一份期望 %d —— 说明同名撞车没解决",
				r, got[i], secondMap[r], firstMap[r])
		}
	}
}

func mustAlias(t *testing.T, data []byte) []byte {
	t.Helper()
	patched, _, err := AliasIconfont(data)
	if err != nil {
		t.Fatal(err)
	}
	return patched
}

// 页面里能用的码位才该出现在 Glyphs 之外；这条盯一下 LoadIconfont 的产物确实把老字体
// 也能过一遍（老字体本来就有别名，不该再造）。
func TestLoadIconfontKeepsExistingAliases(t *testing.T) {
	f, err := LoadIconfont(readFont(t, oldFontPath))
	if err != nil {
		t.Fatal(err)
	}
	if f.Aliases != 0 {
		t.Errorf("老字体本来就有 56xx 别名，却又造了 %d 个", f.Aliases)
	}
	if len(f.Glyphs) != 168 {
		t.Errorf("码位数 %d，期望 168", len(f.Glyphs))
	}
	if f.Family == "" {
		t.Error("家族名是空的")
	}
}
