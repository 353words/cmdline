package main

import (
	"flag"
	"fmt"
	"logs"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s DB_PATH QUERY\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
		os.Exit(1)
	}

	dbPath := flag.Arg(0)
	pathQuery := flag.Arg(1)

	conn, err := sqlx.Open("duckdb", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: connect - %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	filter := logs.Filter{
		Path: pathQuery,
	}

	rows, err := logs.Query(conn, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: query - %v\n", err)
		os.Exit(1)
	}

	for _, row := range rows {
		fmt.Println(row)
	}
}
