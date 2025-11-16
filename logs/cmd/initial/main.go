package main

import (
	"fmt"
	"logs"
	"os"

	"github.com/jmoiron/sqlx"
)

func main() {

	dbPath := os.Args[1]
	pathQuery := os.Args[2]

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
