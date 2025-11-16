package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Record is a log record.
type Record struct {
	Origin     string
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	Size       int
}

// parseLine parses a log line.
func parseLine(line string) (Record, error) {
	// slppp6.intermind.net - - [01/Aug/1995:00:00:10 -0400] "GET /history/skylab/skylab.html HTTP/1.0" 200 1687
	fields := strings.Fields(line)
	var size int
	var err error
	if s := fields[len(fields)-1]; s == "-" {
		size = 0
	} else {
		size, err = strconv.Atoi(s)
		if err != nil {
			return Record{}, err
		}
	}

	code, err := strconv.Atoi(fields[len(fields)-2])
	if err != nil {
		return Record{}, err
	}

	const layout = "[02/Jan/2006:15:04:05 -0700]"
	t, err := time.Parse(layout, fields[3]+" "+fields[4])
	if err != nil {
		return Record{}, err
	}

	r := Record{
		Origin:     fields[0],
		Method:     fields[5][1:], // Remove leading "
		Path:       fields[6],
		StatusCode: code,
		Time:       t.UTC(),
		Size:       size,
	}
	return r, nil
}

// Filter is a query filter.
type Filter struct {
	Path string
	// TODO: Other filter fields
}

// Match returns true if the filter matches r.
func (f Filter) Match(r Record) bool {
	return strings.Contains(r.Path, f.Path)
}

// Query returns logs from r that match filter.
func Query(r io.Reader, filter Filter) ([]Record, error) {
	var result []Record
	s := bufio.NewScanner(r)
	lNum := 0
	for s.Scan() {
		lNum++
		r, err := parseLine(s.Text())
		if err != nil {
			return nil, fmt.Errorf("%d: %w", lNum, err)
		}

		if filter.Match(r) {
			result = append(result, r)
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s QUERY LOG_FILE\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
		os.Exit(1)
	}

	pathQuery := flag.Arg(0)
	fileName := flag.Arg(1)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	filter := Filter{
		Path: pathQuery,
	}

	records, err := Query(file, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: query - %v\n", err)
		os.Exit(1)
	}

	for _, r := range records {
		fmt.Println(r)
	}
}
