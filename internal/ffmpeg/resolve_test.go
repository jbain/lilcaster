package ffmpeg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lilcaster/internal/config"
)

func TestResolve_Passthrough(t *testing.T) {
	e := config.Endpoint{Path: "rtmp://ingest.example.com/abc123/streamkey"}
	re, err := Resolve(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if re.Path != e.Path {
		t.Errorf("path: got %q, want %q", re.Path, e.Path)
	}
	if re.Original != e {
		t.Error("Original not preserved")
	}
}

func TestResolve_ScriptAbsolute(t *testing.T) {
	// Uses the checked-in testdata script via an absolute path.
	abs, err := filepath.Abs("testdata/echo.sh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(abs, 0755); err != nil {
		t.Fatal(err)
	}

	e := config.Endpoint{Path: "script://" + abs}
	re, err := Resolve(e)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if re.Path != "/resolved/path.mp4" {
		t.Errorf("path: got %q, want %q", re.Path, "/resolved/path.mp4")
	}
	if re.Original != e {
		t.Error("Original not preserved")
	}
}

func TestResolve_ScriptRelative(t *testing.T) {
	// Go tests run from the package directory, so testdata/echo.sh is reachable.
	if err := os.Chmod("testdata/echo.sh", 0755); err != nil {
		t.Fatal(err)
	}

	e := config.Endpoint{Path: "script://testdata/echo.sh"}
	re, err := Resolve(e)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if re.Path != "/resolved/path.mp4" {
		t.Errorf("path: got %q, want %q", re.Path, "/resolved/path.mp4")
	}
}

func TestResolve_AVFoundation(t *testing.T) {
	e := config.Endpoint{Path: "avfoundation://0:0"}
	_, err := Resolve(e)
	if err == nil {
		t.Fatal("expected error for avfoundation")
	}
	if !strings.Contains(err.Error(), "avfoundation") {
		t.Errorf("error should mention 'avfoundation': %v", err)
	}
}
