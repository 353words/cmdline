package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	Origin     string
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	Size       int
}

func parseTime(s string) (time.Time, error) {
	const layout = "[02/Jan/2006:15:04:05 -0700]"
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// slppp6.intermind.net - - [01/Aug/1995:00:00:10 -0400] "GET /history/skylab/skylab.html HTTP/1.0" 200 1687
func parseLine(line string) (Record, error) {
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

	t, err := parseTime(fields[3] + " " + fields[4])
	if err != nil {
		return Record{}, err
	}

	r := Record{
		Origin:     fields[0],
		Method:     fields[5][1:], // Remove leading "
		Path:       fields[6],
		StatusCode: code,
		Time:       t,
		Size:       size,
	}
	return r, nil
}

// zcat http.log.gz | go run . > logs.json
func main() {
	s := bufio.NewScanner(os.Stdin)
	lNum := 0
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()
	for s.Scan() {
		lNum++
		r, err := parseLine(s.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%d: %q (%v)\n", lNum, s.Text(), err)
			os.Exit(1)
		}

		w.Write([]string{
			r.Time.Format(time.RFC3339),
			r.Origin,
			r.Method,
			r.Path,
			fmt.Sprintf("%d", r.StatusCode),
			fmt.Sprintf("%d", r.Size),
		})
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1.0)
	}
}
