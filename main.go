package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: samfilter <read_ids.txt> < input.sam > output.sam")
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("usage: samfilter <read_ids.txt> < input.sam > output.sam")
		fmt.Println()
		fmt.Println("Filter a SAM file to only include reads matching the given ID list.")
		fmt.Println("Reads a SAM file from stdin and writes filtered output to stdout.")
		fmt.Println("Header lines (starting with @) are always passed through.")
		os.Exit(0)
	}

	ids, err := readIDs(os.Args[1])
	if err != nil {
		log.Fatalf("error reading id list: %v", err)
	}

	if err := filterSAM(os.Stdin, os.Stdout, ids); err != nil {
		log.Fatalf("error filtering SAM: %v", err)
	}
}

// filterSAM reads SAM-formatted data from r, writes matching records to w.
// Header lines (starting with @) are always included.
// Data lines are included only if their QNAME (first field) is in ids.
// ids must be sorted.
func filterSAM(r io.Reader, w io.Writer, ids []string) error {
	scanner := bufio.NewScanner(r)
	// SAM files can have very long lines; set a 10MB buffer.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == '@' {
			fmt.Fprintln(bw, line)
			continue
		}
		qname, _, _ := strings.Cut(line, "\t")
		if qname == "" {
			continue
		}
		if i := sort.SearchStrings(ids, qname); i < len(ids) && ids[i] == qname {
			fmt.Fprintln(bw, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	return bw.Flush()
}

// readIDs reads a list of read IDs from a text file
// and returns a sorted slice of unique IDs.
func readIDs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	var ids []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ids = append(ids, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	if len(ids) == 0 {
		return nil, errors.New("id list is empty")
	}

	sort.Strings(ids)
	// Deduplicate
	uniq := ids[:1]
	for _, id := range ids[1:] {
		if id != uniq[len(uniq)-1] {
			uniq = append(uniq, id)
		}
	}

	return uniq, nil
}
