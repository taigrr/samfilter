package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

// This small utility will help extract entries
// from a SAM file and output them to a new file
// based on a given list of read ids.
// It expects a list of read ids as a text file as the only argument
// and a source SAM file on stdin.
// It will output the filtered SAM file to stdout.
func main() {
	if len(os.Args) != 2 {
		log.Fatalf("please provide a list of read ids as the only argument (see help)")
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("usage: samfilter <read_ids.txt> < input.sam > output.sam")
		os.Exit(0)
	}
	idList := os.Args[1]
	samFile := os.Stdin
	outFile := os.Stdout
	ids := readIds(idList)

	scanner := bufio.NewScanner(samFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '@' {
			fmt.Fprintln(outFile, line)
			continue
		}
		fields := strings.Fields(line)
		index := sort.SearchStrings(ids, fields[0])
		if index < len(ids) && ids[index] == fields[0] {
			fmt.Fprintln(outFile, line)
		}
	}
}

// readIds reads a list of read ids from a text file
// and returns a slice of sorted, unique ids.
func readIds(idList string) []string {
	idFile, err := os.Open(idList)
	if err != nil {
		log.Fatalf("error opening id list: %v", err)
	}
	defer idFile.Close()
	var ids []string
	scanner := bufio.NewScanner(idFile)
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading id list: %v", err)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		log.Fatalf("id list is empty")
	}
	uniques := []string{ids[0]}
	for i := range ids {
		if i > 0 && ids[i-1] != ids[i] {
			uniques = append(uniques, ids[i])
		}
	}
	return uniques
}
