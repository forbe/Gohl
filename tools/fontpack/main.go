package main

// fontpack 把一支图标字体补好 56xx 别名后写进 resources.zip：页面里一行 @font-face
// 就能用，不需要业务代码配合。
//
//	go run ./tools/fontpack                       # 默认的 testdata/iconfont.new.ttf
//	go run ./tools/fontpack 别的.ttf              # 换成自己从 iconfont.cn 下的字体
//	go run ./tools/fontpack 别的.ttf 我的.zip     # 指定要改的 zip
//
// 要在老项目里换掉内置图标，只要重跑这条命令再提交 zip；业务侧代码不用动。

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/forbe/gohl"
)

const (
	defaultFont = "testdata/iconfont.new.ttf"
	defaultZip  = "resources.zip"
)

type entry struct {
	name string
	mode os.FileMode
	data []byte
}

func main() {
	log.SetFlags(0)
	src, dst := defaultFont, defaultZip
	if len(os.Args) > 1 {
		src = os.Args[1]
	}
	if len(os.Args) > 2 {
		dst = os.Args[2]
	}

	raw, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("读字体 %s 失败: %v", src, err)
	}
	patched, aliases, err := gohl.AliasIconfont(raw)
	if err != nil {
		log.Fatalf("补别名失败: %v", err)
	}
	list, err := gohl.GlyphsOf(patched)
	if err != nil {
		log.Fatalf("补完别名反而读不出码位: %v", err)
	}
	family, err := gohl.FontFamily(patched)
	if err != nil {
		log.Fatalf("读不出家族名: %v", err)
	}

	// zip 里的条目名就用源文件名，resources:// 后面跟的是它。
	name := filepath.Base(src)
	usable, pua := 0, 0
	for _, g := range list {
		if g.Codepoint >= 0xE000 && g.Codepoint <= 0xF8FF {
			pua++
		} else {
			usable++
		}
	}

	if err := pack(dst, name, patched); err != nil {
		log.Fatalf("写 %s 失败: %v", dst, err)
	}

	fmt.Printf("%s → %s（条目名 %s）\n", src, dst, name)
	fmt.Printf("新造别名 %d 个；码位共 %d，页面上能用 %d 个，PUA 原件 %d 个（画不出来，别用）\n",
		aliases, len(list), usable, pua)
	fmt.Printf("CSS：@font-face{ font-family:'%s'; src:url('resources://%s'); }\n", family, name)
}

func pack(zipPath, name string, data []byte) error {
	var entries []entry
	raw, err := os.ReadFile(zipPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		r, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
		if err != nil {
			return fmt.Errorf("读现有 zip: %w", err)
		}
		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}
			entries = append(entries, entry{name: f.Name, mode: f.Mode(), data: b})
		}
	}

	replaced := false
	for i := range entries {
		if entries[i].name == name {
			entries[i].data = data
			replaced = true
		}
	}
	if !replaced {
		entries = append(entries, entry{name: name, mode: 0444, data: data})
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		hdr.SetMode(e.mode) // Modified 留零值：1980-01-01，重跑一次字节不变
		fw, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := fw.Write(e.data); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return os.WriteFile(zipPath, buf.Bytes(), 0644)
}
