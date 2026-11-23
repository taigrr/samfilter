package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkFilterSAMSmallInput(b *testing.B) {
	const input = "@HD\tVN:1.6\nread1\t0\tchr1\t100\t60\t4M\t*\t0\t0\tACGT\t*\n"
	ids := []string{"read1"}
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		if err := filterSAM(strings.NewReader(input), io.Discard, ids); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadIDsSmallInput(b *testing.B) {
	path := filepath.Join(b.TempDir(), "ids.txt")
	if err := os.WriteFile(path, []byte("read1\nread2\n"), 0o600); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := readIDs(path); err != nil {
			b.Fatal(err)
		}
	}
}
