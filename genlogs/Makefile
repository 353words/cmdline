db:
	rm -f logs.db logs.csv
	zcat http.log.gz | go run . > logs.csv
	cat schema.sql | duckdb logs.db
	echo "COPY logs FROM 'logs.csv';" | duckdb logs.db
	rm logs.csv

