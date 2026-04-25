// samfilter extracts entries from a SAM file based on a list of read IDs.
//
// It reads a SAM file from stdin, filters lines whose read ID appears in the
// provided ID list file, and writes matching entries to stdout. Header lines
// (starting with @) are always passed through.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "samfilter <read_ids.txt> < input.sam > output.sam",
		Short:   "Filter SAM file entries by read ID list",
		Long:    "Reads a SAM file from stdin, keeps only entries whose read ID appears in the provided ID list file, and writes to stdout. Header lines (@) are always preserved.",
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE:    run,
	}

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ids, err := readIDs(args[0])
	if err != nil {
		return err
	}

	return filterSAM(os.Stdin, os.Stdout, ids)
}

// maxLineSize is the maximum line length the scanner will accept.
// SAM records can be very long (long reads, large CIGAR strings), so
// we use 10 MB to match what the README advertises.
const maxLineSize = 10 * 1024 * 1024

// filterSAM reads SAM-formatted lines from r and writes matching entries to w.
// Header lines (starting with @) are always passed through. Alignment lines are
// included only if their read ID (first field) appears in the sorted ids slice.
func filterSAM(r io.Reader, w io.Writer, ids []string) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, maxLineSize), maxLineSize)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '@' {
			fmt.Fprintln(w, line)
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		index := sort.SearchStrings(ids, fields[0])
		if index < len(ids) && ids[index] == fields[0] {
			fmt.Fprintln(w, line)
		}
	}
	return scanner.Err()
}

// readIDs reads a list of read IDs from a text file
// and returns a sorted slice of unique IDs.
func readIDs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening ID list: %w", err)
	}
	defer f.Close()

	var ids []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ids = append(ids, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading ID list: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("ID list is empty")
	}

	sort.Strings(ids)

	// Deduplicate
	uniques := ids[:1]
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1] {
			uniques = append(uniques, ids[i])
		}
	}
	return uniques, nil
}

func init() {
	log.SetFlags(0)
}
