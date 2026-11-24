package internal

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xanygo/anygo/xcfg"
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
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return err
		}
		fp, err := downloadFile(item.URL, fileDir)
		if err != nil {
			return err
		}
		out := filepath.Join(fileDir, item.Name)
		if err := xfs.Rename(fp, out); err != nil {
			return err
		}
	}
	return nil
}

type FileConfig struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	ArchivePath string `yaml:"archive_path"` // 在 zip 中的文件路径
	Integrity   string `yaml:"integrity"`    // 该 url 对应的 sri 值
}

type ArchiveConfig struct {
	Type            string `yaml:"type"`
	URL             string `yaml:"url"`
	Filename        string `yaml:"filename"`
	SHA256          string `yaml:"sha256"`
	StripComponents uint   `yaml:"strip_components"`
	// Include     []string `yaml:"include"`
}

func (ac ArchiveConfig) download(dir string) error {
	return downloadArchiveFile(&ac, dir)
}

func (ac ArchiveConfig) relName() string {
	return path.Join("archive", ac.Filename)
}

type AssetConfig struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Source      string   `yaml:"source"`
	File        string   `yaml:"file"`
	URL         string   `yaml:"url"`
	Integrity   string   `yaml:"integrity"`
	CrossOrigin string   `yaml:"cross_origin"`
	Media       string   `yaml:"media"`
	Defer       bool     `yaml:"defer"`
	Async       bool     `yaml:"async"`
	DependsOn   []string `yaml:"depends_on"`
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
	for _, item := range s.Files {
		if item.Name == a.File {
			a.URL = item.URL
			a.Integrity = item.Integrity
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
	for _, item := range s.Files {
		if item.Name == a.File {
			if item.ArchivePath != "" {
				a.URL = path.Join(urlPrefix, "b", s.Name, item.ArchivePath)
			} else {
				a.URL = path.Join(urlPrefix, "a", s.Name, item.Name)
			}
			a.Integrity = item.Integrity
			break
		}
	}
	return a
}

func (a AssetConfig) GoString() string {
	const tpl = `{
		Name:"{{ .Name }}",
		Type:"{{ .Type }}",
		URL:"{{ .URL }}",
{{ if .Integrity -}} Integrity:"{{ .Integrity }}",{{ end }}
{{ if .CrossOrigin -}} CrossOrigin:"{{ .CrossOrigin }}",{{ end }}
{{ if .Media -}} Media: "{{ .Media }}", {{ end }}
{{ if .Defer -}} Defer:{{ .Defer }},  {{ end }}
{{ if .Async -}} Async: {{ .Async }},  {{ end }}
{{ if .DependsOn -}} DependsOn: {{ printf "%#v" .DependsOn }},  {{ end }}
		}`
	str, _ := xstr.RenderTemplate(tpl, a)
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
	if cfg.Output.Mode != "cdn" && cfg.Output.Mode != "embed" && cfg.Output.Mode != "both" {
		return errors.New("output.mode must be cdn, embed or both")
	}
	if cfg.Output.Dir == "" {
		return errors.New("output.dir is required")
	}
	if cfg.Output.Package == "" {
		cfg.Output.Package = "webassets"
	}
	if !isGoIdentifier(cfg.Output.Package) {
		return fmt.Errorf("invalid output.package %q", cfg.Output.Package)
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

	sources := make(map[string]SourceConfig)
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
				return fmt.Errorf("source %q file %q: path is required when archive is configured", s.Name, f.Name)
			}
			if f.ArchivePath != "" && !safeRelativeSlashPath(f.ArchivePath) {
				return fmt.Errorf("source %q file %q: unsafe path %q", s.Name, f.Name, f.ArchivePath)
			}
		}
		if s.Archive != nil {
			if strings.ToLower(s.Archive.Type) != "zip" {
				return fmt.Errorf("source %q: only zip archives are supported", s.Name)
			}
			if s.Archive.URL == "" {
				return fmt.Errorf("source %q: archive.url and archive.include are required", s.Name)
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
		if a.Type != "js" && a.Type != "css" {
			return fmt.Errorf("asset %q: type must be js or css", a.Name)
		}
		if a.Async && a.Defer {
			return fmt.Errorf("asset %q: async and defer cannot both be true", a.Name)
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
		for _, dep := range a.DependsOn {
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
		deps[a.Name] = a.DependsOn
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
