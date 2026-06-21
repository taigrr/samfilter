package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterSAM(t *testing.T) {
	input := strings.Join([]string{
		"@HD\tVN:1.6\tSO:coordinate",
		"@SQ\tSN:chr1\tLN:248956422",
		"read1\t0\tchr1\t100\t60\t50M\t*\t0\t0\tACGT\t*",
		"read2\t0\tchr1\t200\t60\t50M\t*\t0\t0\tACGT\t*",
		"read3\t0\tchr1\t300\t60\t50M\t*\t0\t0\tACGT\t*",
		"read4\t0\tchr1\t400\t60\t50M\t*\t0\t0\tACGT\t*",
	}, "\n") + "\n"

	ids := []string{"read1", "read3"}

	var buf bytes.Buffer
	if err := filterSAM(strings.NewReader(input), &buf, ids); err != nil {
		t.Fatalf("filterSAM: %v", err)
	}

	got := buf.String()
	// Headers should pass through
	if !strings.Contains(got, "@HD") {
		t.Error("missing @HD header")
	}
	if !strings.Contains(got, "@SQ") {
		t.Error("missing @SQ header")
	}
	// Matching reads
	if !strings.Contains(got, "read1\t") {
		t.Error("missing read1")
	}
	if !strings.Contains(got, "read3\t") {
		t.Error("missing read3")
	}
	// Non-matching reads
	if strings.Contains(got, "read2\t") {
		t.Error("should not contain read2")
	}
	if strings.Contains(got, "read4\t") {
		t.Error("should not contain read4")
	}
}

func TestFilterSAM_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	if err := filterSAM(strings.NewReader(""), &buf, []string{"read1"}); err != nil {
		t.Fatalf("filterSAM: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestFilterSAM_EmptyLines(t *testing.T) {
	input := "@HD\tVN:1.6\n\nread1\t0\tchr1\t100\t60\t50M\t*\t0\t0\tACGT\t*\n\n"
	var buf bytes.Buffer
	if err := filterSAM(strings.NewReader(input), &buf, []string{"read1"}); err != nil {
		t.Fatalf("filterSAM: %v", err)
	}
	if !strings.Contains(buf.String(), "read1") {
		t.Error("missing read1")
	}
}

func TestFilterSAM_IgnoresSpaceSeparatedNonSAMLines(t *testing.T) {
	input := strings.Join([]string{
		"read1 extra tokens",
		"read1\t0\tchr1\t100\t60\t50M\t*\t0\t0\tACGT\t*",
	}, "\n") + "\n"

	var buf bytes.Buffer
	if err := filterSAM(strings.NewReader(input), &buf, []string{"read1"}); err != nil {
		t.Fatalf("filterSAM: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "read1 extra tokens") {
		t.Error("unexpectedly included malformed space-separated line")
	}
	if !strings.Contains(got, "read1\t0\tchr1") {
		t.Error("missing valid SAM record")
	}
}

func TestReadIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ids.txt")
	if err := os.WriteFile(path, []byte("banana\napple\ncherry\napple\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ids, err := readIDs(path)
	if err != nil {
		t.Fatalf("readIDs: %v", err)
	}

	want := []string{"apple", "banana", "cherry"}
	if len(ids) != len(want) {
		t.Fatalf("got %d ids, want %d", len(ids), len(want))
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestReadIDs_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ids.txt")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := readIDs(path)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestReadIDs_SkipsCommentLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ids.txt")
	if err := os.WriteFile(path, []byte("# keep only these reads\nfoo\n  # another comment\nbar\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ids, err := readIDs(path)
	if err != nil {
		t.Fatalf("readIDs: %v", err)
	}

	want := []string{"bar", "foo"}
	if len(ids) != len(want) {
		t.Fatalf("got %d ids, want %d", len(ids), len(want))
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestReadIDs_WhitespaceLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ids.txt")
	if err := os.WriteFile(path, []byte("  foo  \n\n  bar  \n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ids, err := readIDs(path)
	if err != nil {
		t.Fatalf("readIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d ids, want 2", len(ids))
	}
}

func TestReadIDs_LongLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ids.txt")
	longID := strings.Repeat("read", 20000) // 80KB
	if err := os.WriteFile(path, []byte(longID+"\nshort\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ids, err := readIDs(path)
	if err != nil {
		t.Fatalf("readIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d ids, want 2", len(ids))
	}
	foundLongID := false
	for _, id := range ids {
		if id == longID {
			foundLongID = true
			break
		}
	}
	if !foundLongID {
		t.Errorf("long ID was not preserved")
	}
}

func TestFilterSAM_LongLine(t *testing.T) {
	// Simulate a SAM record longer than the default 64KB scanner buffer.
	longSeq := strings.Repeat("ACGT", 20000) // 80KB
	input := "@HD\tVN:1.6\n" +
		"read1\t0\tchr1\t100\t60\t50M\t*\t0\t0\t" + longSeq + "\t*\n" +
		"read2\t0\tchr1\t200\t60\t50M\t*\t0\t0\tACGT\t*\n"

	var buf bytes.Buffer
	if err := filterSAM(strings.NewReader(input), &buf, []string{"read1"}); err != nil {
		t.Fatalf("filterSAM: %v", err)
	}
	if !strings.Contains(buf.String(), "read1") {
		t.Error("missing read1 with long sequence")
	}
	if strings.Contains(buf.String(), "read2") {
		t.Error("should not contain read2")
	}
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestFilterSAM_ReadError(t *testing.T) {
	wantErr := errors.New("boom")

	err := filterSAM(failingReader{err: wantErr}, io.Discard, []string{"read1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reading SAM input") {
		t.Fatalf("expected SAM read context, got %v", err)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error %v, got %v", wantErr, err)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestFilterSAM_HeaderWriteError(t *testing.T) {
	wantErr := errors.New("write failed")

	err := filterSAM(strings.NewReader("@HD\tVN:1.6\n"), failingWriter{err: wantErr}, []string{"read1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "writing SAM header") {
		t.Fatalf("expected SAM header write context, got %v", err)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error %v, got %v", wantErr, err)
	}
}

func TestFilterSAM_RecordWriteError(t *testing.T) {
	wantErr := errors.New("write failed")
	input := "read1\t0\tchr1\t100\t60\t50M\t*\t0\t0\tACGT\t*\n"

	err := filterSAM(strings.NewReader(input), failingWriter{err: wantErr}, []string{"read1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "writing SAM record") {
		t.Fatalf("expected SAM record write context, got %v", err)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error %v, got %v", wantErr, err)
	}
}

func TestFilterSAM_LineTooLong(t *testing.T) {
	tooLongRead := strings.Repeat("A", maxLineSize+1)
	input := tooLongRead + "\t0\tchr1\t100\t60\t50M\t*\t0\t0\tACGT\t*\n"

	err := filterSAM(strings.NewReader(input), io.Discard, []string{tooLongRead})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "reading SAM input: line exceeds") {
		t.Fatalf("expected line-too-long context, got %v", err)
	}
	if !errors.Is(err, bufio.ErrTooLong) {
		t.Fatalf("expected bufio.ErrTooLong, got %v", err)
	}
}

func TestReadIDs_NotFound(t *testing.T) {
	_, err := readIDs("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
