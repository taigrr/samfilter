# samfilter

A fast, simple utility to filter [SAM](https://samtools.github.io/hts-specs/SAMv1.pdf) files by read ID.

## Usage

```bash
samfilter <read_ids.txt> < input.sam > output.sam
```

- `read_ids.txt` — a text file with one read ID per line
- Reads a SAM file from stdin
- Writes filtered SAM output to stdout
- Header lines (`@` prefix) are always passed through
- Only alignment records with a QNAME matching one of the provided IDs are included

## Install

```bash
go install github.com/taigrr/samfilter@latest
```

## Example

```bash
# Extract specific reads from a SAM file
echo -e "read_001\nread_042\nread_999" > my_ids.txt
samtools view -h input.bam | samfilter my_ids.txt > filtered.sam
```

## Features

- Sorted binary search for fast ID lookup
- Automatic deduplication of input IDs
- Handles whitespace and empty lines in ID files
- 10 MB line buffer for long SAM records
- Buffered I/O for efficient streaming

## License

[0BSD](LICENSE)
