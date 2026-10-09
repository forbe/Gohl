package gohl

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

//go:embed resources.zip
var resourcesZip embed.FS

var (
	resourcesDir string

	resourcesOnce   sync.Once
	resourcesReader *zip.Reader
	resourcesErr    error
)

// extractResources 把内置的 resources.zip 释放到 %APPDATA%/gohl。
//
// 逐个文件比大小，而不是打一个「已释放」标记：框架升级往 zip 里加了新文件（比如
// iconfont.new.ttf）时，老用户的目录里有标记、却没有那个文件，标记式释放就永远补不上。
func extractResources() {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.TempDir()
	}

	resourcesDir = filepath.Join(appData, "gohl")
	os.MkdirAll(resourcesDir, 0755)

	reader, err := zipReader()
	if err != nil {
		fmt.Printf("解压 resources.zip 失败: %v\n", err)
		return
	}

	for _, file := range reader.File {
		dstPath := filepath.Join(resourcesDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(dstPath, file.Mode())
			continue
		}

		if st, err := os.Stat(dstPath); err == nil && st.Size() == int64(file.UncompressedSize) {
			continue // 磁盘上已经是这一版
		}

		if err := writeExtracted(dstPath, file); err != nil {
			fmt.Printf("释放资源 %s 失败: %v\n", file.Name, err)
			continue
		}
		fmt.Printf("释放资源: %s\n", dstPath)
	}
}

func writeExtracted(dstPath string, file *zip.File) error {
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	os.MkdirAll(filepath.Dir(dstPath), 0755)
	// zip 里的 dll 带着只读位，覆盖前先让目标可写，否则 O_TRUNC 直接被拒。
	if st, err := os.Stat(dstPath); err == nil && st.Mode()&0200 == 0 {
		os.Chmod(dstPath, st.Mode()|0600)
	}

	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// zipReader 打开内嵌的 resources.zip，目录只解一次。
func zipReader() (*zip.Reader, error) {
	resourcesOnce.Do(func() {
		data, err := resourcesZip.ReadFile("resources.zip")
		if err != nil {
			resourcesErr = fmt.Errorf("读取嵌入的 resources.zip 失败: %w", err)
			return
		}
		resourcesReader, resourcesErr = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	})
	return resourcesReader, resourcesErr
}

// ReadResource 从内置 resources.zip 里按名字读出资源字节（如 "iconfont.new.ttf"），
// 不用等它被释放到磁盘，也不用自己再嵌一份。
func ReadResource(name string) ([]byte, error) {
	reader, err := zipReader()
	if err != nil {
		return nil, err
	}
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("resources.zip 里没有 %s", name)
}

func GetResourcesDir() string {
	return resourcesDir
}
