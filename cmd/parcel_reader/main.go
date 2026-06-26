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
	// 1. Open CSV and read header
	header, reader, cleanup, err := getCSVReader("data/parcels.csv")
	if err != nil {
		log.Fatalf("failed to get CSV reader: %v", err)
	}
	defer cleanup()

	// 2. Open SQLite
	db, err := sql.Open("sqlite", "parcels.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Create Table
	if err := createTable(db, header); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 4. Insert Rows
	count, err := insertParcels(db, header, reader)
	if err != nil {
		log.Fatalf("failed to insert parcels: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into parcels.db\n", count)
}

func getCSVReader(filePath string) (header []string, reader *csv.Reader, cleanup func(), err error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open CSV: %w", err)
	}

	reader = csv.NewReader(csvFile)

	header, err = reader.Read()
	if err != nil {
		csvFile.Close()
		return nil, nil, nil, fmt.Errorf("failed to read header: %w", err)
	}

	cleanup = func() {
		csvFile.Close()
	}

	return header, reader, cleanup, nil
}

func createTable(db *sql.DB, header []string) error {
	var columnDefs []string
	for _, col := range header {
		colName := strings.Trim(col, "\"")
		columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" TEXT", colName))
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS parcels (%s);", strings.Join(columnDefs, ", "))
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func insertParcels(db *sql.DB, header []string, reader *csv.Reader) (int, error) {
	var quotedCols []string
	var placeholders []string

	for _, col := range header {
		colName := strings.Trim(col, "\"")
		quotedCols = append(quotedCols, fmt.Sprintf("\"%s\"", colName))
		placeholders = append(placeholders, "?")
	}

	insertSQL := fmt.Sprintf("INSERT INTO parcels (%s) VALUES (%s);", strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "))

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

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

		if len(record) != len(header) {
			log.Printf("record length mismatch: expected %d, got %d", len(header), len(record))
			continue
		}

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
	return count, nil
}