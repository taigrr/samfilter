package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseVersion(t *testing.T) {
	const releaseVersion = "1.2.3"
	binary := filepath.Join(t.TempDir(), "samfilter")
	build := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w -X main.version="+releaseVersion, "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build release binary: %v\n%s", err, output)
	}
	for _, flag := range []string{"--version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			output, err := exec.Command(binary, flag).CombinedOutput()
			if err != nil {
				t.Fatalf("version command: %v\n%s", err, output)
			}
			if got := strings.TrimSpace(string(output)); got != "samfilter version "+releaseVersion {
				t.Fatalf("got %q, want release version %q", got, releaseVersion)
			}
		})
	}
}
