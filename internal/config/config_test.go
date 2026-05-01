package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func loadFromString(s string) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal([]byte(s), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func TestLoad_ValidFixture(t *testing.T) {
	cfg, err := Load("testdata/valid.yml")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Scenarios) != 1 {
		t.Fatalf("scenario count: got %d, want 1", len(cfg.Scenarios))
	}
	sc := cfg.Scenarios[0]

	if sc.Name != "simple_stream" {
		t.Errorf("name: got %q, want %q", sc.Name, "simple_stream")
	}
	if len(sc.Sources) != 1 || sc.Sources[0].Path != "/Users/jbain/video_assets/big_buck_bunny_1080p_stereo.avi" {
		t.Errorf("sources: %+v", sc.Sources)
	}
	if len(sc.Sinks) != 1 || sc.Sinks[0].Path != "rtmp://127.0.0.1:1935/live/test" {
		t.Errorf("sinks: %+v", sc.Sinks)
	}
	if sc.Loop != -1 {
		t.Errorf("loop: got %d, want -1", sc.Loop)
	}
	if len(sc.Filters) != 4 {
		t.Fatalf("filter count: got %d, want 4", len(sc.Filters))
	}

	scale, ok := sc.Filters[0].Filter.(*ScaleFilter)
	if !ok {
		t.Fatalf("filter[0]: expected *ScaleFilter, got %T", sc.Filters[0].Filter)
	}
	if scale.Width != "-2" || scale.Height != "480" {
		t.Errorf("scale: width=%q height=%q", scale.Width, scale.Height)
	}

	if _, ok := sc.Filters[1].Filter.(*TimestampFilter); !ok {
		t.Fatalf("filter[1]: expected *TimestampFilter, got %T", sc.Filters[1].Filter)
	}

	custom, ok := sc.Filters[2].Filter.(*CustomFilter)
	if !ok {
		t.Fatalf("filter[2]: expected *CustomFilter, got %T", sc.Filters[2].Filter)
	}
	if custom.String != "eq=brightness=0.1" {
		t.Errorf("custom string: got %q", custom.String)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("testdata/nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_ScaleMissingHeight(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: scale
        width: "100"
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for scale missing height")
	}
	if !strings.Contains(err.Error(), "height") {
		t.Errorf("error should mention 'height': %v", err)
	}
}

func TestLoad_ScaleMissingBoth(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: scale
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for scale with no dimensions")
	}
}

func TestLoad_CustomMissingString(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: custom
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for custom missing string")
	}
}

func TestLoad_TimestampExtraField(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: timestamp
        extra: forbidden
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for timestamp with extra field")
	}
}

func TestLoad_PhraseFilter(t *testing.T) {
	const doc = `
scenarios:
  - name: ok
    sources: []
    sinks: []
    filters:
      - type: phrase
    loop: 0
`
	cfg, err := loadFromString(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.Scenarios[0].Filters[0].Filter.(*PhraseFilter); !ok {
		t.Errorf("filter[0]: expected *PhraseFilter, got %T", cfg.Scenarios[0].Filters[0].Filter)
	}
}

func TestLoad_PhraseExtraField(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: phrase
        extra: forbidden
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for phrase with extra field")
	}
}

func TestLoad_EndpointArgs(t *testing.T) {
	const doc = `
scenarios:
  - name: ok
    sources:
      - path: script://get-source.sh
        args:
          - "--env"
          - "production"
    sinks:
      - path: rtmp://example.com/live
    loop: 0
`
	cfg, err := loadFromString(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := cfg.Scenarios[0].Sources[0]
	if src.Path != "script://get-source.sh" {
		t.Errorf("path: got %q", src.Path)
	}
	if len(src.Args) != 2 || src.Args[0] != "--env" || src.Args[1] != "production" {
		t.Errorf("args: got %v", src.Args)
	}
}

func TestLoad_UnknownFilterType(t *testing.T) {
	const doc = `
scenarios:
  - name: bad
    sources: []
    sinks: []
    filters:
      - type: magic
    loop: 0
`
	_, err := loadFromString(doc)
	if err == nil {
		t.Fatal("expected error for unknown filter type")
	}
	if !strings.Contains(err.Error(), "magic") {
		t.Errorf("error should name the bad type: %v", err)
	}
}
