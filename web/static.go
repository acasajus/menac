package web

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/acasajus/menac/web/binstatic"
)

type StaticMgr struct {
	JsFiles  []string
	CssFiles []string
}

func processRequiresContents(reqPath string, ext string) []string {
	path := filepath.Dir(reqPath)
	dirs := strings.Split(path, string(os.PathSeparator))
	prefix := path
	for i, d := range dirs {
		if d == "static" {
			prefix = filepath.Join(dirs[i+1:]...)
			break
		}
	}

	reqData, err := binstatic.Asset(reqPath)
	if err != nil {
	}
	assets := make([]string, 0, 1)
	liner := bufio.NewReader(bytes.NewBuffer(reqData))
	dotext := "." + ext
	for line, _ := liner.ReadString('\n'); line != ""; line, _ = liner.ReadString('\n') {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		fields := strings.Split(strings.TrimSpace(line), " ")
		if len(fields) > 1 {
			switch {
			case fields[0] == "require":
				if filepath.Ext(fields[1]) != dotext {
					fields[1] = fields[1] + dotext
				}
				assetPath := filepath.Join(path, fields[1])
				if fi, err := binstatic.AssetInfo(assetPath); err == nil && !fi.IsDir() {
					assets = append(assets, filepath.Join(prefix, fields[1]))
				} else {
				}
			case fields[0] == "require_tree":
				assetPath := filepath.Join(path, fields[1])
				binstatic.Walk(assetPath, func(subPath string, info os.FileInfo, err error) error {
					if info.IsDir() || filepath.Ext(subPath) != dotext {
						return nil
					}
					relPath, _ := filepath.Rel(assetPath, subPath)
					ap := filepath.Join(prefix, relPath)
					assets = append(assets, ap)
					return nil
				})
			}
		}
	}
	return assets
}

func findStaticAssets(basePath string, ext string) ([]string, error) {
	assets := []string{}
	return assets, binstatic.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) != "requires" {
			return nil
		}
		entries := processRequiresContents(path, ext)
		assets = append(assets, entries...)
		return nil
	})
}

func NewStaticMgr() *StaticMgr {
	sm := &StaticMgr{}
	return sm
}