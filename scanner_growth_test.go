package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScannersGrowToLineLimit(t *testing.T) {
	// Leave room for the newline within Scanner's maximum buffer size.
	line := strings.Repeat("A", maxLineSize-2)
	t.Run("SAM header", func(t *testing.T) {
		input := "@" + line + "\n"
		var output bytes.Buffer
		if err := filterSAM(strings.NewReader(input), &output, nil); err != nil {
			t.Fatal(err)
		}
		if output.String() != input {
			t.Fatal("header changed while growing the scanner buffer")
		}
	})
	t.Run("ID list", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "ids.txt")
		if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		ids, err := readIDs(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 1 || ids[0] != line {
			t.Fatal("ID changed while growing the scanner buffer")
		}
	})
}
