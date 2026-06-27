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
	loadParcels()
	loadParcelGeo()
}

func loadParcels() {
	// 1. Open CSV and read header
	header, reader, csvFile, err := getCSVReader("data/parcels.csv")
	if err != nil {
		log.Fatalf("failed to get CSV reader: %v", err)
	}
	defer csvFile.Close()

	// 2. Open SQLite
	db, err := sql.Open("sqlite", "data/parcels.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Create Table
	if err := createTable(db, "parcels", header); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 4. Insert Rows
	count, err := insertParcels(db, "parcels", header, reader)
	if err != nil {
		log.Fatalf("failed to insert parcels: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into data/parcels.db\n", count)
}

func loadParcelGeo() {
	// 1. Open CSV and read header
	header, reader, csvFile, err := getCSVReader("data/parcel_geo.csv")
	if err != nil {
		log.Fatalf("failed to get CSV reader: %v", err)
	}
	defer csvFile.Close()

	// 2. Open SQLite
	db, err := sql.Open("sqlite", "data/parcels.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Create Table
	if err := createTable(db, "parcel_geo", header); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 4. Insert Rows
	count, err := insertParcels(db, "parcel_geo", header, reader)
	if err != nil {
		log.Fatalf("failed to insert parcel geo: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into data/parcels.db (parcel_geo)\n", count)
}

func getCSVReader(filePath string) (header []string, reader *csv.Reader, file *os.File, err error) {
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

	return header, reader, csvFile, nil
}

func createTable(db *sql.DB, tableName string, header []string) error {
	var columnDefs []string
	for _, col := range header {
		colName := strings.Trim(col, "\"")
		colName = strings.ReplaceAll(colName, "\"", "\"\"")
		columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" TEXT", colName))
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columnDefs, ", "))
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func insertParcels(db *sql.DB, tableName string, header []string, reader *csv.Reader) (int, error) {
	var quotedCols []string
	var placeholders []string

	for _, col := range header {
		colName := strings.Trim(col, "\"")
		colName = strings.ReplaceAll(colName, "\"", "\"\"")
		quotedCols = append(quotedCols, fmt.Sprintf("\"%s\"", colName))
		placeholders = append(placeholders, "?")
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(quotedCols, ", "), strings.Join(placeholders, ", "))

	count := 0
	txCount := 0
	const batchSize = 10000

	// Helper to start a transaction and prepare the statement
	startTx := func() (*sql.Tx, *sql.Stmt, error) {
		tx, err := db.Begin()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		stmt, err := tx.Prepare(insertSQL)
		if err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to prepare statement: %w", err)
		}
		return tx, stmt, nil
	}

	tx, stmt, err := startTx()
	if err != nil {
		return 0, err
	}

	// Ensure cleanup of the last transaction/statement
	defer func() {
		if tx != nil {
			stmt.Close()
			tx.Rollback() // Will be a no-op if committed
		}
	}()

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
		txCount++

		if txCount >= batchSize {
			if err := tx.Commit(); err != nil {
				return count, fmt.Errorf("failed to commit batch: %w", err)
			}
			stmt.Close()

			tx, stmt, err = startTx()
			if err != nil {
				return count, err
			}
			txCount = 0
			fmt.Printf("inserted %d records (committed batch)...\n", count)
		} else if count%1000 == 0 {
			fmt.Printf("inserted %d records...\n", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("failed to commit final batch: %w", err)
	}

	return count, nil
}
