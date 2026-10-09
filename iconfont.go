package gohl

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"unicode/utf16"
)

// iconfont 支持：把 iconfont.cn 下载的字体变成 HTMLayout 能直接用的样子。
//
// 两件事必须在这里做，使用者的页面才不会出现 '?'：
//  1. PUA 别名。HTMLayout 走 GBK 解码，PUA（E000-F8FF）过不了它，画出来是 '?'；而
//     iconfont.cn 现在下载的字体整支都挂在 PUA。同一批字形另开到 56xx-59xx（GBK 解得开
//     的一段）就能用，内置那份 iconfont.new.ttf 已经在 resources.zip 里预先补好了。
//  2. 唯一家族名。HTMLayout/GDI 是按「家族名」挑 face 的，不是按我们给的 URL。同名第二份
//     注册了也不生效（实测 GDI 仍回第一份的字形号），所以交付前把自报名换成唯一名。

const (
	puaLo      = 0xE000 // HTMLayout 画不出来的整段
	puaHi      = 0xF8FF
	puaAliasLo = 0xE600 // 这段内沿用 cp-0x9000 的固定公式
	puaAliasHi = 0xE9FF
	aliasOff   = 0x9000 // E61D -> 561D
	aliasLo    = 0x5600 // 别名只能落在这段：GBK 解得开，1024 个坑
	aliasHi    = 0x59FF
	tagTTCF    = 0x74746366
)

// Glyph 是字体里一个可用码位：Name 取自 post 表的 v2.0 名字数组（iconfont 生成的 ttf
// 都在那里带上 class 名），拿不到就留空。
type Glyph struct {
	Codepoint rune
	GlyphID   uint16
	Name      string
}

// Iconfont 是一份可以直接交付给引擎的图标字体。
type Iconfont struct {
	Family  string  // CSS 里要写的 font-family（已换成唯一名）
	Data    []byte  // 交给引擎的字体字节，PUA 字形已补上 56xx 别名
	Aliases int     // 这份字体里新造出来的别名个数
	Glyphs  []Glyph // 全部码位，按码位升序；PUA 原件也在里面但画不出来
}

var iconfontSeq int64

// AliasIconfont 给字体补 56xx 别名：返回重写 cmap 后的字节和造出来的别名个数。
// 已经挂过别名的（比如内置那份）会得到 0 个、原字节照还。E6xx-E9xx 仍按老公式
// cp-0x9000（和内置那份对得上），其余 PUA 从 0x5600 往后逐个占空位——iconfont 现在
// 会把图标发到 ECxx/EDxx，只认老公式的话一份 359 字的字体只能救回 87 个。
func AliasIconfont(data []byte) ([]byte, int, error) {
	if len(data) >= 4 && binary.BigEndian.Uint32(data[:4]) == tagTTCF {
		return data, 0, nil // ttc 集合改 cmap 要动整包，图标字体不会是 ttc，跳过
	}
	orig, err := readCmap(data)
	if err != nil {
		return data, 0, err
	}
	byCp := make(map[rune]uint16, len(orig)*2)
	for _, g := range orig {
		byCp[g.r] = g.gid
	}

	// has 记录哪些字形在 56xx-59xx 里已经有坑位了：内置那份本来就有别名，一个都不该再造。
	has := make(map[uint16]bool, len(orig))
	for _, g := range orig {
		if g.r >= aliasLo && g.r <= aliasHi {
			has[g.gid] = true
		}
	}

	var pending []glyph
	n := 0
	for _, g := range orig {
		if g.r < puaLo || g.r > puaHi || has[g.gid] {
			continue
		}
		if g.r >= puaAliasLo && g.r <= puaAliasHi {
			alias := g.r - aliasOff
			if alias >= aliasLo && alias <= aliasHi {
				if _, taken := byCp[alias]; !taken {
					byCp[alias] = g.gid // 老公式，和内置那份一模一样
					has[g.gid] = true
					n++
					continue
				}
			}
		}
		pending = append(pending, g)
	}
	// 剩下的没有固定公式可用，从 0x5600 起挨个占没人用的坑。
	next := aliasLo
	for _, g := range pending {
		if has[g.gid] {
			continue // 同一个字形被两个码位共用，一个别名就够
		}
		for next <= aliasHi {
			alias := rune(next)
			next++
			if _, taken := byCp[alias]; taken {
				continue
			}
			byCp[alias] = g.gid
			has[g.gid] = true
			n++
			break
		}
	}
	if n == 0 {
		return data, 0, nil
	}
	cmap, err := buildCmap(byCp)
	if err != nil {
		return data, 0, err
	}
	patched, err := replaceCmap(data, cmap)
	if err != nil {
		return data, 0, err
	}
	return patched, n, nil
}

// LoadIconfont 是给自己带字体（从 iconfont.cn 下回来的一份 ttf）用的唯一入口：
// 补 PUA 别名 + 换成唯一家族名 + 解出码位表，一次做完。
func LoadIconfont(data []byte) (*Iconfont, error) {
	seq := int(atomic.AddInt64(&iconfontSeq, 1))
	patched, aliases, err := AliasIconfont(data)
	if err != nil {
		return nil, err
	}
	family := fmt.Sprintf("gohl-icon-%d", seq)
	if base, berr := familyOf(data); berr == nil {
		family = fmt.Sprintf("%s-%d", sanitizeFamily(base), seq)
	}
	out, err := renameFont(patched, family)
	if err != nil {
		return nil, err
	}
	glyphs, err := GlyphsOf(out)
	if err != nil {
		return nil, err
	}
	return &Iconfont{Family: family, Data: out, Aliases: aliases, Glyphs: glyphs}, nil
}

// BundledIconfont 取出 resources.zip 里内置的那支 iconfont.new.ttf 并处理好。内置那份
// 已预先补过别名，这里的别名数会是 0；拿到 Glyphs 是为了在 Go 里码出常量表。
func BundledIconfont() (*Iconfont, error) {
	data, err := ReadResource("iconfont.new.ttf")
	if err != nil {
		return nil, err
	}
	return LoadIconfont(data)
}

// GlyphsOf 只读码位表，不改字体字节。名字取自 post 表，iconfont.cn 生成的字体里就是
// 那个 class 名；页面里能用的是 Codepoint，直接把它写进 HTML。
func GlyphsOf(data []byte) ([]Glyph, error) {
	list, err := readCmap(data)
	if err != nil {
		return nil, err
	}
	out := make([]Glyph, len(list))
	for i, g := range list {
		out[i] = Glyph{Codepoint: g.r, GlyphID: g.gid, Name: g.name}
	}
	return out, nil
}

// FontFamily 读出字体自报的家族名。@font-face 的 font-family 必须和它一字不差，写成
// 别名会静默退回系统字体（规则见 skill）。
func FontFamily(data []byte) (string, error) {
	return familyOf(data)
}

// Serve 把这份字体注册到自定义 scheme 上，返回 HTML 里 @font-face 要写的 URI。
// URI 里带上唯一家族名，引擎按 URI 缓存才不会把上一份字体端出来。同一个 scheme 再注册
// 会覆盖。
func (f *Iconfont) Serve(scheme string) string {
	uri := scheme + "://" + f.Family + ".ttf"
	data := f.Data
	RegisterResourceLoader(scheme, func(string) ([]byte, uint32, bool) {
		// 不自己报类型：GetResourceDataType 对 .ttf 返回 0（引擎的 AUTO），实测图标能出。
		return data, GetResourceDataType(uri), true
	})
	return uri
}

// CSS 返回一行可以直接放进 <style> 的 @font-face 规则。
func (f *Iconfont) CSS(uri string) string {
	return fmt.Sprintf("@font-face{ font-family:'%s'; src:url('%s'); }", f.Family, uri)
}

func sanitizeFamily(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "gohl-icon"
	}
	return strings.NewReplacer(" ", "-", "'", "", "\"", "", "/", "-").Replace(s)
}

// ---------- 码位表读取 ----------

// glyph 是内部用的码位记录，读 cmap 时带字形号（造别名要靠它对齐）。
type glyph struct {
	r    rune
	gid  uint16
	name string
}

// readCmap 返回按码位升序的全部 Unicode 码位映射。支持 ttf、CFF 的 otf 和 ttc 集合
// （集合取第 0 个字体的那张表）。
//
// cmap 子表可能是 format 4（BMP，16 位）或 format 12（全 Unicode，32 位）。优先读
// format 12；只有 format 4 时才回过去读它。
func readCmap(font []byte) ([]glyph, error) {
	if len(font) < 16 {
		return nil, fmt.Errorf("字体太小(%d 字节)，不是完整字体文件", len(font))
	}
	base := 0
	if binary.BigEndian.Uint32(font[:4]) == tagTTCF { // 'ttcf'：版本号 + 字数 + 偏移数组，取第 0 个
		if n := binary.BigEndian.Uint32(font[8:12]); n < 1 {
			return nil, fmt.Errorf("ttc 声明 %d 个字体，读不出第 0 个的偏移", n)
		}
		base = int(binary.BigEndian.Uint32(font[12:16]))
		if base+12 > len(font) {
			return nil, fmt.Errorf("ttc 第 0 个字体的偏移 %d 越过文件末尾", base)
		}
	}
	switch tag := binary.BigEndian.Uint32(font[base : base+4]); tag {
	case 0x00010000, 0x74727565, 0x4f54544f: // \x00\x01\x00\x00 / 'true' / 'OTTO'
	default:
		return nil, fmt.Errorf("sfnt 版本标记 %08x 不认（只支持 ttf/otf/ttc）", tag)
	}
	num := int(binary.BigEndian.Uint16(font[base+4 : base+6]))
	if base+12+num*16 > len(font) {
		return nil, fmt.Errorf("表目录声明 %d 张表，超出文件长度", num)
	}

	var cmap, post, maxp int
	var postLen, maxpLen uint32
	for i := 0; i < num; i++ {
		off := base + 12 + i*16
		tag := string(font[off : off+4])
		o := int(binary.BigEndian.Uint32(font[off+8 : off+12]))
		l := binary.BigEndian.Uint32(font[off+12 : off+16])
		if o+int(l) > len(font) {
			return nil, fmt.Errorf("表 %s 声明偏移 %d 长度 %d，越过文件末尾", tag, o, l)
		}
		switch tag {
		case "cmap":
			cmap = o
		case "post":
			post, postLen = o, l
		case "maxp":
			maxp, maxpLen = o, l
		}
	}
	if cmap == 0 {
		return nil, fmt.Errorf("字体里没有 cmap 表，读不出码位")
	}
	numGlyphs := uint16(0)
	if maxp != 0 {
		if maxpLen < 6 {
			return nil, fmt.Errorf("maxp 表长度 %d 不够", maxpLen)
		}
		numGlyphs = binary.BigEndian.Uint16(font[maxp+4 : maxp+6])
	}

	names := readPostNames(font, post, postLen, numGlyphs)

	sub, err := pickCmapSubtable(font, cmap)
	if err != nil {
		return nil, err
	}
	glyphs := make([]glyph, 0, len(sub))
	for _, g := range sub {
		if g.gid == 0 {
			continue // gid 0 是 .notdef，画出来就是空白框
		}
		if int(g.gid) < len(names) {
			g.name = names[g.gid]
		}
		glyphs = append(glyphs, g)
	}
	if len(glyphs) == 0 {
		return nil, fmt.Errorf("cmap 里没有任何非 .notdef 的 Unicode 映射")
	}
	sort.Slice(glyphs, func(i, j int) bool { return glyphs[i].r < glyphs[j].r })
	return glyphs, nil
}

// pickCmapSubtable 在目录里挑一张 Unicode 表：先 format 12，再 format 4。
func pickCmapSubtable(font []byte, cmap int) ([]glyph, error) {
	if cmap+4 > len(font) {
		return nil, fmt.Errorf("cmap 表头越界")
	}
	n := int(binary.BigEndian.Uint16(font[cmap+2 : cmap+4]))
	var fmt4, fmt12 int
	for i := 0; i < n; i++ {
		o := cmap + 4 + i*8
		if o+8 > len(font) {
			return nil, fmt.Errorf("cmap 目录第 %d 项越界", i)
		}
		pid := binary.BigEndian.Uint16(font[o : o+2])
		eid := binary.BigEndian.Uint16(font[o+2 : o+4])
		so := cmap + int(binary.BigEndian.Uint32(font[o+4:o+8]))
		if pid != 0 && pid != 3 {
			continue // 只认 Unicode 和 Windows 两个平台
		}
		if so+2 > len(font) {
			continue
		}
		switch fmt := binary.BigEndian.Uint16(font[so : so+2]); fmt {
		case 12:
			if fmt12 == 0 {
				fmt12 = so
			}
		case 4:
			if pid == 3 && eid == 1 && fmt4 == 0 {
				fmt4 = so
			} else if fmt4 == 0 {
				fmt4 = so
			}
		}
	}
	if fmt12 != 0 {
		return parseFormat12(font, fmt12)
	}
	if fmt4 != 0 {
		return parseFormat4(font, fmt4)
	}
	return nil, fmt.Errorf("cmap 里没有 format 4/12 子表，读不出码位")
}

func parseFormat12(font []byte, so int) ([]glyph, error) {
	if so+16 > len(font) {
		return nil, fmt.Errorf("format 12 表头越界")
	}
	groups := int(binary.BigEndian.Uint32(font[so+12 : so+16]))
	if so+16+groups*12 > len(font) {
		return nil, fmt.Errorf("format 12 声明 %d 组，越过文件末尾", groups)
	}
	var out []glyph
	for i := 0; i < groups; i++ {
		o := so + 16 + i*12
		start := binary.BigEndian.Uint32(font[o : o+4])
		end := binary.BigEndian.Uint32(font[o+4 : o+8])
		gid := binary.BigEndian.Uint32(font[o+8 : o+12])
		if end < start || end-start > 0xFFFF {
			return nil, fmt.Errorf("format 12 第 %d 组区间 %04X..%04X 不合理", i, start, end)
		}
		for c := start; c <= end; c++ {
			if c > 0x10FFFF {
				break
			}
			out = append(out, glyph{r: rune(c), gid: uint16(gid + (c - start))})
		}
	}
	return out, nil
}

func parseFormat4(font []byte, so int) ([]glyph, error) {
	if so+14 > len(font) {
		return nil, fmt.Errorf("format 4 表头越界")
	}
	segs := int(binary.BigEndian.Uint16(font[so+6:so+8])) / 2
	endAt := so + 14
	startAt := endAt + segs*2 + 2
	deltaAt := startAt + segs*2
	rangeAt := deltaAt + segs*2
	if rangeAt+segs*2 > len(font) {
		return nil, fmt.Errorf("format 4 声明 %d 段，越过文件末尾", segs)
	}
	var out []glyph
	for i := 0; i < segs; i++ {
		end := int(binary.BigEndian.Uint16(font[endAt+i*2 : endAt+i*2+2]))
		start := int(binary.BigEndian.Uint16(font[startAt+i*2 : startAt+i*2+2]))
		delta := int(binary.BigEndian.Uint16(font[deltaAt+i*2 : deltaAt+i*2+2]))
		ro := int(binary.BigEndian.Uint16(font[rangeAt+i*2 : rangeAt+i*2+2]))
		if i == segs-1 {
			continue // 最后一段是 end=0xFFFF 的哨兵
		}
		if end < start {
			continue
		}
		for c := start; c <= end; c++ {
			gid := 0
			if ro == 0 {
				gid = (c + delta) & 0xFFFF
			} else {
				po := rangeAt + i*2 + ro + (c-start)*2
				if po+2 > len(font) {
					return nil, fmt.Errorf("format 4 第 %d 段的字形数组指针越界", i)
				}
				g := int(binary.BigEndian.Uint16(font[po : po+2]))
				if g != 0 {
					gid = (g + delta) & 0xFFFF
				}
			}
			if c < 0x20 || gid == 0 {
				continue
			}
			out = append(out, glyph{r: rune(c), gid: uint16(gid)})
		}
	}
	return out, nil
}

// readPostNames 取 post v2.0 里的自定义字形名。非 2.0 版本、或长度对不上，就整体
// 返回空表：名字是锦上添花，码位才是正事，不能因为 post 读不出就白屏。
func readPostNames(font []byte, post int, postLen uint32, numGlyphs uint16) []string {
	names := make([]string, numGlyphs)
	if post == 0 || post+32 > len(font) {
		return names
	}
	if binary.BigEndian.Uint32(font[post:post+4]) != 0x00020000 {
		return names
	}
	ng := int(binary.BigEndian.Uint16(font[post+32 : post+34]))
	if ng != int(numGlyphs) {
		return names
	}
	idxAt := post + 34
	if idxAt+ng*2 > len(font) {
		return names
	}
	strAt := idxAt + ng*2
	end := post + int(postLen)
	if end > len(font) {
		end = len(font)
	}
	for i := 0; i < ng; i++ {
		id := int(binary.BigEndian.Uint16(font[idxAt+i*2 : idxAt+i*2+2]))
		if id < 258 {
			continue // 标准 Mac 名，这套字体里用不上
		}
		p := strAt
		for k := 258; k < id; k++ {
			if p >= end {
				return names
			}
			p += 1 + int(font[p])
		}
		if p >= end {
			return names
		}
		l := int(font[p])
		if p+1+l > end {
			return names
		}
		names[i] = string(font[p+1 : p+1+l])
	}
	return names
}

// ---------- cmap 重写 ----------

// buildCmap 用给定的码位表重造一张 cmap：BMP 部分进 format 4（Windows/GDI 主读的
// 就是它），有非 BMP 的码位再补一张 format 12。
func buildCmap(byCp map[rune]uint16) ([]byte, error) {
	var bmp, all, nonBMP []int
	for r := range byCp {
		all = append(all, int(r))
		if r > 0xFFFF {
			nonBMP = append(nonBMP, int(r))
			continue
		}
		if r >= 0xD800 && r <= 0xDFFF {
			continue // 代理区不映射
		}
		bmp = append(bmp, int(r))
	}
	sort.Ints(all)
	sort.Ints(bmp)

	type sub struct {
		pid, eid uint16
		data     []byte
	}
	f4, err := buildFormat4(bmp, byCp)
	if err != nil {
		return nil, err
	}
	subs := []sub{{3, 1, f4}, {0, 3, f4}}
	if len(nonBMP) > 0 {
		f12 := buildFormat12(all, byCp)
		subs = append(subs, sub{3, 10, f12}, sub{0, 4, f12})
	}

	hdr := 4 + 8*len(subs)
	offsets := make([]int, len(subs))
	pos := hdr
	for i, s := range subs {
		offsets[i] = pos
		pos += len(s.data)
		if pos%4 != 0 {
			pos += 4 - pos%4
		}
	}

	out := make([]byte, 0, pos)
	head := make([]byte, 4)
	binary.BigEndian.PutUint16(head, 0)
	binary.BigEndian.PutUint16(head[2:], uint16(len(subs)))
	out = append(out, head...)
	for i, s := range subs {
		e := make([]byte, 8)
		binary.BigEndian.PutUint16(e, s.pid)
		binary.BigEndian.PutUint16(e[2:], s.eid)
		binary.BigEndian.PutUint32(e[4:], uint32(offsets[i]))
		out = append(out, e...)
	}
	for i, s := range subs {
		for len(out) < offsets[i] {
			out = append(out, 0)
		}
		out = append(out, s.data...)
	}
	return pad4(out), nil
}

func pad4(b []byte) []byte {
	for len(b)%4 != 0 {
		b = append(b, 0)
	}
	return b
}

type fmt4Seg struct {
	start, end int
}

func buildFormat4(cps []int, byCp map[rune]uint16) ([]byte, error) {
	var segs []fmt4Seg
	for i := 0; i < len(cps); {
		j := i
		for j+1 < len(cps) && cps[j+1] == cps[j]+1 {
			j++
		}
		segs = append(segs, fmt4Seg{cps[i], cps[j]})
		i = j + 1
	}
	segs = append(segs, fmt4Seg{0xFFFF, 0xFFFF}) // 哨兵段，规范要求的收尾
	n := len(segs)

	arrays := make([][]uint16, n)
	total := 0
	for i, s := range segs {
		if s.start == 0xFFFF {
			continue
		}
		g := make([]uint16, 0, s.end-s.start+1)
		for c := s.start; c <= s.end; c++ {
			g = append(g, byCp[rune(c)])
		}
		arrays[i] = g
		total += len(g)
	}

	// 14 字节表头 + 四份段数组(8n) + 保留 pad(2) + 字形数组
	body := 16 + 8*n + 2*total
	if body > 0xFFFF {
		return nil, fmt.Errorf("format 4 装不下 %d 个码位（表长 %d 字节，上限 65535）", len(cps), body)
	}
	b := make([]byte, 0, body)
	u16 := func(v int) { b = binary.BigEndian.AppendUint16(b, uint16(v)) }
	u16(4)
	u16(body)
	u16(0)
	u16(2 * n)
	es := 0
	for k := n; k > 1; k >>= 1 {
		es++
	}
	sr := (1 << es) * 2
	u16(sr)
	u16(es)
	u16(2*n - sr)
	for _, s := range segs {
		u16(s.end)
	}
	u16(0)
	for _, s := range segs {
		u16(s.start)
	}
	for _, s := range segs {
		if s.start == 0xFFFF {
			u16(1) // 哨兵段用 delta=1，把 0xFFFF 算成 gid 0
		} else {
			u16(0) // 其余走 idRangeOffset，delta 归零
		}
	}
	prefix := 0
	for i, s := range segs {
		if s.start == 0xFFFF {
			u16(0)
			continue
		}
		u16(2*(n-i) + prefix)
		prefix += len(arrays[i]) * 2
	}
	for _, g := range arrays {
		for _, v := range g {
			u16(int(v))
		}
	}
	if len(b) != body {
		return nil, fmt.Errorf("format 4 编出 %d 字节，和声明的 %d 不符", len(b), body)
	}
	return b, nil
}

func buildFormat12(cps []int, byCp map[rune]uint16) []byte {
	type grp struct{ start, end, gid uint32 }
	var gs []grp
	for _, c := range cps {
		gid := uint32(byCp[rune(c)])
		if n := len(gs); n > 0 {
			last := &gs[n-1]
			if uint32(c) == last.end+1 && gid == last.gid+(uint32(c)-last.start) {
				last.end = uint32(c)
				continue
			}
		}
		gs = append(gs, grp{uint32(c), uint32(c), gid})
	}
	length := 16 + 12*len(gs)
	b := make([]byte, 0, length)
	u16 := func(v int) { b = binary.BigEndian.AppendUint16(b, uint16(v)) }
	u32 := func(v uint32) { b = binary.BigEndian.AppendUint32(b, v) }
	u16(12)
	u16(0)
	u32(uint32(length))
	u32(uint32(len(gs)))
	for _, g := range gs {
		u32(g.start)
		u32(g.end)
		u32(g.gid)
	}
	return b
}

// replaceCmap 把新的 cmap 交给通用的换表走法：追加到文件末尾，只改目录项里的偏移和长度。
// 旧 cmap 留在原地变成没人引用的死字节，其余表的位置不动，所以不用重排整个文件。
func replaceCmap(font, cmap []byte) ([]byte, error) {
	return replaceTable(font, sfntBase(font), "cmap", cmap)
}

// ---------- name 表（家族名） ----------

// nameID 里这几条共同参与家族名/全称匹配，一并重写；16/17/18 是 Typographic 系列，
// Windows 有它们就优先用它们，留着会让改名失效。
var overriddenNames = map[uint16]bool{1: true, 2: true, 3: true, 4: true, 6: true, 16: true, 17: true, 18: true}

type nameRec struct {
	pid, eid, lid, nid uint16
	data               []byte
}

// renameFont 返回换了自报名字的字体字节；只动 name 表，cmap 和字形一律原样。
func renameFont(font []byte, name string) ([]byte, error) {
	base := sfntBase(font)
	at, _, _, err := tableAt(font, base, "name")
	if err != nil {
		return nil, err
	}
	recs, err := parseNames(font, at)
	if err != nil {
		return nil, err
	}
	var kept []nameRec
	for _, r := range recs {
		if !overriddenNames[r.nid] {
			kept = append(kept, r)
		}
	}
	psName := strings.ReplaceAll(name, " ", "-")
	uniq := name + "-gohl"
	// 平台 0（Unicode）和 3（Windows）用 UTF-16BE，平台 1（Mac）用单字节 ASCII：
	// 三套都补上，GDI/DirectWrite/老引擎各认各的那一套。
	for _, w := range []struct {
		pid, eid, lid uint16
		enc           func(string) []byte
	}{
		{0, 3, 0, utf16BE},
		{3, 1, 0x409, utf16BE},
		{1, 0, 0, asciiBytes},
	} {
		for _, e := range []struct {
			nid  uint16
			text string
		}{
			{1, name}, {2, "Regular"}, {3, uniq}, {4, name}, {6, psName},
		} {
			kept = append(kept, nameRec{w.pid, w.eid, w.lid, e.nid, w.enc(e.text)})
		}
	}
	tbl := buildNameTable(kept)
	return replaceTable(font, base, "name", tbl)
}

// familyOf 读出字体自报的家族名（name 表 ID=1，Windows 平台那条）。
func familyOf(font []byte) (string, error) {
	at, _, _, err := tableAt(font, sfntBase(font), "name")
	if err != nil {
		return "", err
	}
	recs, err := parseNames(font, at)
	if err != nil {
		return "", err
	}
	for _, r := range recs {
		if r.nid == 1 && r.pid == 3 && len(r.data) >= 4 {
			u := make([]uint16, len(r.data)/2)
			for i := range u {
				u[i] = binary.BigEndian.Uint16(r.data[i*2:])
			}
			return string(utf16.Decode(u)), nil
		}
	}
	return "", fmt.Errorf("name 表里没有 Windows 平台的家族名")
}

func utf16BE(s string) []byte {
	out := make([]byte, 0, 2*len(s))
	for _, u := range utf16.Encode([]rune(s)) {
		out = binary.BigEndian.AppendUint16(out, u)
	}
	return out
}

func asciiBytes(s string) []byte {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			r = '?'
		}
		b = append(b, byte(r))
	}
	return b
}

func parseNames(font []byte, at int) ([]nameRec, error) {
	if at+6 > len(font) {
		return nil, fmt.Errorf("name 表头越界")
	}
	if binary.BigEndian.Uint16(font[at:at+2]) != 0 {
		return nil, fmt.Errorf("name 表 format 是 %d，只认 0", binary.BigEndian.Uint16(font[at:at+2]))
	}
	count := int(binary.BigEndian.Uint16(font[at+2 : at+4]))
	so := at + int(binary.BigEndian.Uint16(font[at+4:at+6]))
	if so <= at+6 || so > len(font) {
		return nil, fmt.Errorf("name 表字符串区偏移 %d 不合理", so-at)
	}
	out := make([]nameRec, 0, count)
	for i := 0; i < count; i++ {
		o := at + 6 + i*12
		if o+12 > len(font) {
			return nil, fmt.Errorf("name 表第 %d 条记录越界", i)
		}
		l := int(binary.BigEndian.Uint16(font[o+8 : o+10]))
		s := so + int(binary.BigEndian.Uint16(font[o+10:o+12]))
		if s+l > len(font) {
			return nil, fmt.Errorf("name 表第 %d 条字符串越界", i)
		}
		out = append(out, nameRec{
			pid:  binary.BigEndian.Uint16(font[o : o+2]),
			eid:  binary.BigEndian.Uint16(font[o+2 : o+4]),
			lid:  binary.BigEndian.Uint16(font[o+4 : o+6]),
			nid:  binary.BigEndian.Uint16(font[o+6 : o+8]),
			data: font[s : s+l],
		})
	}
	return out, nil
}

func buildNameTable(recs []nameRec) []byte {
	stringAt := 6 + 12*len(recs)
	b := make([]byte, 0, stringAt+256)
	b = binary.BigEndian.AppendUint16(b, 0)
	b = binary.BigEndian.AppendUint16(b, uint16(len(recs)))
	b = binary.BigEndian.AppendUint16(b, uint16(stringAt))
	off := 0
	for _, r := range recs {
		b = binary.BigEndian.AppendUint16(b, r.pid)
		b = binary.BigEndian.AppendUint16(b, r.eid)
		b = binary.BigEndian.AppendUint16(b, r.lid)
		b = binary.BigEndian.AppendUint16(b, r.nid)
		b = binary.BigEndian.AppendUint16(b, uint16(len(r.data)))
		b = binary.BigEndian.AppendUint16(b, uint16(off))
		off += len(r.data)
	}
	for _, r := range recs {
		b = append(b, r.data...)
	}
	return b
}

// ---------- sfnt 表目录 ----------

// sfntBase 处理 ttc：集合里第 0 个字体的表目录不在文件开头，而在偏移表里给的地址。
func sfntBase(font []byte) int {
	if len(font) >= 16 && binary.BigEndian.Uint32(font[:4]) == tagTTCF {
		return int(binary.BigEndian.Uint32(font[12:16]))
	}
	return 0
}

// tableAt 在（可能带基址的）表目录里找一张表，返回它的偏移、长度和目录项地址。
func tableAt(font []byte, base int, tag string) (off, size int, dirAt int, err error) {
	if base+12 > len(font) {
		return 0, 0, 0, fmt.Errorf("偏移 %d 处的表目录越界", base)
	}
	num := int(binary.BigEndian.Uint16(font[base+4 : base+6]))
	if base+12+num*16 > len(font) {
		return 0, 0, 0, fmt.Errorf("表目录声明 %d 张表，超出文件长度", num)
	}
	for i := 0; i < num; i++ {
		d := base + 12 + i*16
		if string(font[d:d+4]) != tag {
			continue
		}
		o := int(binary.BigEndian.Uint32(font[d+8 : d+12]))
		l := int(binary.BigEndian.Uint32(font[d+12 : d+16]))
		if o+l > len(font) {
			return 0, 0, 0, fmt.Errorf("表 %s 声明偏移 %d 长度 %d，越过文件末尾", tag, o, l)
		}
		return o, l, d, nil
	}
	return 0, 0, 0, fmt.Errorf("表目录里没有 %s，没法替换", tag)
}

// replaceTable 把新表追加到文件末尾，只改目录项里的偏移和长度：其余表位置不动，
// 旧表留在原地变成没人引用的死字节。
func replaceTable(font []byte, base int, tag string, data []byte) ([]byte, error) {
	_, _, dirAt, err := tableAt(font, base, tag)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(font)+len(data)+4)
	out = append(out, font...)
	out = pad4(out)
	binary.BigEndian.PutUint32(out[dirAt+8:], uint32(len(out)))
	binary.BigEndian.PutUint32(out[dirAt+12:], uint32(len(data)))
	return append(out, data...), nil
}
