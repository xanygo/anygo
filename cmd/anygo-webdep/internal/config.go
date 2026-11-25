package internal

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xenc/xcompress"
	"github.com/xanygo/anygo/xhtml/webdep"
	"github.com/xanygo/anygo/xio/xfs"
	"github.com/xanygo/anygo/xstr"
)

func loadConfig(filename string) (Config, error) {
	var cfg Config
	err := xcfg.Parse(filename, &cfg)
	return cfg, err
}

type Config struct {
	Output  OutputConfig   `yaml:"output"`
	Sources []SourceConfig `yaml:"sources"`
	Assets  []AssetConfig  `yaml:"assets"`
}

func (c Config) hasAssetDir() bool {
	for _, item := range c.Sources {
		if item.Archive == nil {
			return true
		}
	}
	return false
}

type OutputConfig struct {
	Mode          string `yaml:"mode"`
	Dir           string `yaml:"dir"`
	Package       string `yaml:"package"`
	HandlerPrefix string `yaml:"handler_prefix"`
}

func (oc OutputConfig) cdnGoFile() string {
	return filepath.Join(oc.Dir, "asset_cdn.go")
}

func (oc OutputConfig) cleanCDNFile() {
	os.Remove(oc.cdnGoFile())
}

func (oc OutputConfig) cleanEmbedFile() {
	os.Remove(oc.embedGoFile())
	os.RemoveAll(filepath.Join(oc.Dir, "archive"))
	os.RemoveAll(filepath.Join(oc.Dir, "asset"))
}

func (oc OutputConfig) embedGoFile() string {
	return filepath.Join(oc.Dir, "asset_embed.go")
}

type SourceConfig struct {
	Name    string         `yaml:"name"`
	Files   []FileConfig   `yaml:"files"`
	Archive *ArchiveConfig `yaml:"archive"`
}

func (sc SourceConfig) download(dir string) error {
	if sc.Archive != nil {
		return sc.Archive.download(dir)
	}
	return sc.downloadFromCDN(dir)
}

func (sc SourceConfig) downloadFromCDN(dir string) error {
	for _, item := range sc.Files {
		fileDir := filepath.Join(dir, "asset", sc.Name)

		log.Println("start download", item.URL)
		start := time.Now()
		tmpPath, err := download(context.Background(), item.URL, item.integrityToHash())
		cost := time.Since(start)
		log.Println("download", item.URL, "used=", cost.String(), "err=", err)
		if err != nil {
			return err
		}
		outPath := filepath.Join(fileDir, item.Name)
		err = xfs.CopyFile(tmpPath, outPath)
		if err != nil {
			return err
		}
	}
	return nil
}

type FileConfig struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	ArchivePath string `yaml:"archive_path"` // 可选，在 zip 中的文件路径，在有 Archive 时必填
	Integrity   string `yaml:"integrity"`    // 该 url 对应的 sri 值
}

func (f FileConfig) integrityToHash() string {
	if f.Integrity == "" {
		return ""
	}
	arr := strings.SplitN(f.Integrity, "-", 2)
	if len(arr) != 2 {
		return f.Integrity
	}
	bf, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return f.Integrity
	}
	return arr[0] + ":" + hex.EncodeToString(bf)
}

type ArchiveConfig struct {
	Type            string `yaml:"type"`             // 可选，默认 zip
	URL             string `yaml:"url"`              // 必填，zip 文件 url 地址
	Filename        string `yaml:"filename"`         // 可选，文件名，如 bootstrap.zip
	Hash            string `yaml:"hash"`             // 可选，zip 文件的 Hash 值，如 sha256:xxx
	StripComponents uint   `yaml:"strip_components"` // 可选，zip 文件内需要移除的上层目录数
}

func (ac ArchiveConfig) download(dir string) error {
	log.Println("start download", ac.URL)
	start := time.Now()
	tmpPath, err := download(context.Background(), ac.URL, ac.Hash)
	cost := time.Since(start)
	log.Println("download", ac.URL, "used=", cost.String(), "err=", err)
	if err != nil {
		return err
	}

	outPath := filepath.Join(dir, ac.relName())

	if ac.StripComponents > 0 {
		zc := &xcompress.ZipFileStrip{
			Input:           tmpPath,
			Output:          outPath,
			StripComponents: ac.StripComponents,
			Overwrite:       true,
			Filter: func(name string) bool {
				bn := path.Base(name)
				return !strings.HasPrefix(bn, ".")
			},
		}
		return zc.Run()
	}
	return xfs.CopyFile(tmpPath, outPath)
}

func (ac ArchiveConfig) relName() string {
	return path.Join("archive", ac.Filename)
}

type AssetConfig struct {
	Name    string            `yaml:"name"`
	Type    webdep.AssetType  `yaml:"type"`
	Source  string            `yaml:"source"`
	File    string            `yaml:"file"` // 对应的 SourceConfig.Files.Name
	URL     string            `yaml:"url"`
	Attrs   map[string]string `yaml:"attrs"`   // 其他属性
	Depends []string          `yaml:"depends"` // 依赖的 AssetConfig.Name
}

func (a AssetConfig) transCDN(sources []SourceConfig) AssetConfig {
	if a.URL != "" {
		return a
	}
	var s SourceConfig
	for _, item := range sources {
		if a.Source == item.Name {
			s = item
			break
		}
	}
	if a.Attrs == nil {
		a.Attrs = map[string]string{}
	}
	for _, item := range s.Files {
		if item.Name == a.File {
			a.URL = item.URL
			if item.Integrity != "" {
				a.Attrs["integrity"] = item.Integrity
			}
			break
		}
	}
	return a
}

func (a AssetConfig) transEmbed(urlPrefix string, sources []SourceConfig) AssetConfig {
	var s SourceConfig
	for _, item := range sources {
		if a.Source == item.Name {
			s = item
			break
		}
	}
	if a.Attrs == nil {
		a.Attrs = map[string]string{}
	}
	for _, item := range s.Files {
		if item.Name == a.File {
			if item.ArchivePath != "" {
				a.URL = path.Join(urlPrefix, "b", s.Name, item.ArchivePath)
			} else {
				a.URL = path.Join(urlPrefix, "a", s.Name, item.Name)
			}
			if item.Integrity != "" {
				a.Attrs["integrity"] = item.Integrity
			}
			break
		}
	}
	return a
}

func (a AssetConfig) GoString() string {
	const tpl = `{
		Name:"{{ .Name }}",
		 Type:{{ .Type }},
		URL:"{{ .URL }}",
{{ if .Attrs -}} 
 Attrs:map[string]string{
{{ range $k,$v:=.Attrs }}
	{{ printf "%q:%q" $k $v }},
{{ end }}
 },
 {{ end }}
{{ if .Depends -}} 
 Depends: []string{
	{{ range $index,$item:=.Depends }}
		{{ printf "%#v" $item }},
	{{ end }}
},  
 {{ end }}
		}`
	data := map[string]any{
		"Name":    a.Name,
		"URL":     a.URL,
		"Depends": a.Depends,
		"Attrs":   a.Attrs,
	}
	switch a.Type {
	case webdep.AssetCSS:
		data["Type"] = "webdep.AssetCSS"
	case webdep.AssetJS:
		data["Type"] = "webdep.AssetJS"
	case webdep.AssetLink:
		data["Type"] = "webdep.AssetLink"
	default:
		panic("unsupport Type=" + a.Type)
	}
	str, _ := xstr.RenderTemplate(tpl, data)
	return trimLines(str)
}

func trimLines(str string) string {
	arr := strings.Split(str, "\n")
	lines := make([]string, 0, len(arr))
	for _, item := range arr {
		item = strings.TrimSpace(item)
		if item != "" {
			lines = append(lines, item)
		}
	}
	return strings.Join(lines, "\n")
}

func normalizeAndValidate(cfg *Config) error {
	if cfg.Output.Mode == "" {
		cfg.Output.Mode = "cdn"
	}
	switch cfg.Output.Mode {
	case "cdn", "embed":
		// pass
	default:
		return fmt.Errorf("invalid output.mode=%q", cfg.Output.Mode)
	}
	if cfg.Output.Dir == "" {
		return errors.New("output.dir is required")
	}
	if cfg.Output.Package == "" {
		cfg.Output.Package = "webassets"
	}
	if cfg.Output.HandlerPrefix == "" {
		cfg.Output.HandlerPrefix = "/asset/"
	}
	if !strings.HasPrefix(cfg.Output.HandlerPrefix, "/") {
		return errors.New("output.handler.prefix must start with /")
	}
	if !strings.HasSuffix(cfg.Output.HandlerPrefix, "/") {
		cfg.Output.HandlerPrefix += "/"
	}

	sources := make(map[string]SourceConfig, len(cfg.Sources))
	for i := range cfg.Sources {
		s := &cfg.Sources[i]
		if s.Name == "" {
			return fmt.Errorf("sources[%d].name is required", i)
		}
		if _, ok := sources[s.Name]; ok {
			return fmt.Errorf("duplicate source name %q", s.Name)
		}
		fileNames := make(map[string]bool)
		for j := range s.Files {
			f := &s.Files[j]
			if f.Name == "" || f.URL == "" {
				return fmt.Errorf("source %q files[%d]: name and url are required", s.Name, j)
			}
			if fileNames[f.Name] {
				return fmt.Errorf("source %q: duplicate file name %q", s.Name, f.Name)
			}
			fileNames[f.Name] = true
			if s.Archive != nil && f.ArchivePath == "" {
				return fmt.Errorf("source %q file %q: ArchivePath is required when archive is configured", s.Name, f.Name)
			}
			if f.ArchivePath != "" && !safeRelativeSlashPath(f.ArchivePath) {
				return fmt.Errorf("source %q file %q: unsafe ArchivePath %q", s.Name, f.Name, f.ArchivePath)
			}
		}
		if s.Archive != nil {
			if strings.ToLower(s.Archive.Type) != "zip" {
				return fmt.Errorf("source %q: only zip archives are supported", s.Name)
			}
			if s.Archive.URL == "" {
				return fmt.Errorf("source %q: archive.url and archive.include are required", s.Name)
			}
			if s.Archive.Filename == "" {
				s.Archive.Filename = s.Name + ".zip"
			}
		}
		sources[s.Name] = *s
	}

	assetNames := make(map[string]bool)
	for i, a := range cfg.Assets {
		if a.Name == "" {
			return fmt.Errorf("assets[%d].name is required", i)
		}
		if assetNames[a.Name] {
			return fmt.Errorf("duplicate asset name %q", a.Name)
		}
		assetNames[a.Name] = true

		switch a.Type {
		case webdep.AssetCSS, webdep.AssetJS, webdep.AssetLink:
			// pass
		default:
			return fmt.Errorf("asset %q: type must be js or css or link", a.Name)
		}

		if a.URL != "" {
			if a.Source != "" || a.File != "" {
				return fmt.Errorf("asset %q: url cannot be combined with source or file", a.Name)
			}
		} else {
			if a.Source == "" || a.File == "" {
				return fmt.Errorf("asset %q: source and file are required unless url is set", a.Name)
			}
			_, ok := sources[a.Source]
			if !ok {
				return fmt.Errorf("asset %q: unknown source %q", a.Name, a.Source)
			}
			if !safeRelativeSlashPath(a.File) {
				return fmt.Errorf("asset %q: unsafe local file path %q", a.Name, a.File)
			}
		}
	}
	for _, a := range cfg.Assets {
		for _, dep := range a.Depends {
			if !assetNames[dep] {
				return fmt.Errorf("asset %q: unknown dependency %q", a.Name, dep)
			}
		}
	}
	if err := checkDependencyCycles(cfg.Assets); err != nil {
		return err
	}
	return nil
}

func checkDependencyCycles(assets []AssetConfig) error {
	deps := make(map[string][]string, len(assets))
	for _, a := range assets {
		deps[a.Name] = a.Depends
	}
	state := make(map[string]uint8)
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			return fmt.Errorf("dependency cycle detected at asset %q", name)
		case 2:
			return nil
		}
		state[name] = 1
		for _, dep := range deps[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		state[name] = 2
		return nil
	}
	for name := range deps {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}
