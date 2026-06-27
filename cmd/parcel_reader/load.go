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

type RowReader interface {
	Read() ([]string, error)
}

type filteredReader struct {
	source  *csv.Reader
	indices []int
}

func (f *filteredReader) Read() ([]string, error) {
	record, err := f.source.Read()
	if err != nil {
		return nil, err
	}
	filtered := make([]string, len(f.indices))
	for i, idx := range f.indices {
		filtered[i] = record[idx]
	}
	return filtered, nil
}

func loadParcels(csvPath, tableName string, targetCols []string) {
	if err := loadTable(csvPath, tableName, targetCols); err != nil {
		log.Fatalf("failed to load %s: %v", tableName, err)
	}
}

func loadParcelGeo(csvPath, tableName string, targetCols []string) {
	if err := loadTable(csvPath, tableName, targetCols); err != nil {
		log.Fatalf("failed to load %s: %v", tableName, err)
	}
}

func loadCondos(csvPath, tableName string, targetCols []string) {
	if err := loadTable(csvPath, tableName, targetCols); err != nil {
		log.Fatalf("failed to load %s: %v", tableName, err)
	}
}

func loadResidentialBuildings(csvPath, tableName string, targetCols []string) {
	if err := loadTable(csvPath, tableName, targetCols); err != nil {
		log.Fatalf("failed to load %s: %v", tableName, err)
	}
}

func loadApartments(csvPath, tableName string, targetCols []string) {
	if err := loadTable(csvPath, tableName, targetCols); err != nil {
		log.Fatalf("failed to load %s: %v", tableName, err)
	}
}

func loadTable(csvPath, tableName string, targetCols []string) error {
	header, reader, csvFile, err := getCSVReader(csvPath)
	if err != nil {
		return fmt.Errorf("failed to get CSV reader: %w", err)
	}
	defer csvFile.Close()

	db, err := sql.Open("sqlite", "data/parcels.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	indices := make([]int, 0, len(targetCols))
	newHeader := make([]string, 0, len(targetCols))
	for _, target := range targetCols {
		found := false
		for i, h := range header {
			if h == target {
				indices = append(indices, i)
				newHeader = append(newHeader, h)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required column %s not found in CSV", target)
		}
	}

	fReader := &filteredReader{source: reader, indices: indices}

	if err := createTable(db, tableName, newHeader); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	count, err := insertParcels(db, tableName, newHeader, fReader)
	if err != nil {
		return fmt.Errorf("failed to insert %s: %w", tableName, err)
	}

	fmt.Printf("Successfully inserted %d records into %s\n", count, tableName)
	return nil
}

func getCSVReader(filePath string) (header []string, reader *csv.Reader, file *os.File, err error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open CSV: %w", err)
	}

	reader = csv.NewReader(csvFile)
	reader.LazyQuotes = true

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

	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	if _, err := db.Exec(dropTableSQL); err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columnDefs, ", "))
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func insertParcels(db *sql.DB, tableName string, header []string, reader RowReader) (int, error) {
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

	defer func() {
		if tx != nil {
			stmt.Close()
			tx.Rollback()
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
			vals[i] = strings.TrimSpace(v)
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
