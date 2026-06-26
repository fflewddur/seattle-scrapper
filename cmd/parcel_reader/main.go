package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	// 1. Open CSV
	csvFile, err := os.Open("data/parcels.csv")
	if err != nil {
		log.Fatalf("failed to open CSV: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// 2. Read Header
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("failed to read header: %v", err)
	}

	// 3. Open SQLite
	db, err := sql.Open("sqlite", "parcels.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 4. Create Table
	var columnDefs []string
	var quotedCols []string
	var placeholders []string

	for _, col := range header {
		colName := strings.Trim(col, "\"")
		columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" TEXT", colName))
		quotedCols = append(quotedCols, fmt.Sprintf("\"%s\"", colName))
		placeholders = append(placeholders, "?")
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS parcels (%s);", strings.Join(columnDefs, ", "))
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 5. Prepare Insert Statement
	insertSQL := fmt.Sprintf("INSERT INTO parcels (%s) VALUES (%s);", strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "))

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// 6. Read and Insert Rows
	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading record: %v", err)
			continue
		}

		// Ensure record length matches header length
		if len(record) != len(header) {
			log.Printf("record length mismatch: expected %d, got %d", len(header), len(record))
			continue
		}

		// Convert record to []interface{} for stmt.Exec
		vals := make([]any, len(record))
		for i, v := range record {
			vals[i] = v
		}

		_, err = stmt.Exec(vals...)
		if err != nil {
			log.Printf("failed to insert record: %v", err)
			continue
		}
		count++
		if count%1000 == 0 {
			fmt.Printf("inserted %d records...\n", count)
		}
	}

	fmt.Printf("Successfully inserted %d records into parcels.db\n", count)
}
