package buildinfo_test

import (
	"testing"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
)

func TestBuildInfo_Get(t *testing.T) {
	info := buildinfo.Get()
	if info.Version == "" {
		t.Error("expected non-empty Version")
	}
	if info.Go == "" {
		t.Error("expected non-empty Go version")
	}
	if info.OS == "" {
		t.Error("expected non-empty OS")
	}
	if info.Arch == "" {
		t.Error("expected non-empty Arch")
	}
}
