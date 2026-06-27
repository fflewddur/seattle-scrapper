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

func main() {
	loadParcels()
	loadParcelGeo()
	loadCondos()
	loadResidentialBuildings()
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

	// 3. Filter columns
	targetCols := []string{
		"Major",
		"Minor",
		"PropName",
		"PropType",
		"PlatName",
		"PlatLot",
		"PlatBlock",
	}
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
			log.Fatalf("required column %s not found in CSV", target)
		}
	}

	fReader := &filteredReader{
		source:  reader,
		indices: indices,
	}

	// 4. Create Table
	if err := createTable(db, "parcels", newHeader); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 5. Insert Rows
	count, err := insertParcels(db, "parcels", newHeader, fReader)
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

	// 3. Filter columns
	targetCols := []string{
		"Major",
		"Minor",
		"Site Address (per KCA)",
		"Property Name (per KCA)",
		"Property Type (per KCA)",
		"Land Use Code (per KCA)",
		"Detailed Existing Land Use (per KCA)",
		"Site Zip Code (per KCA)",
		"Building Description (per KCA)",
		"Ownership Type",
		"Public Ownership Category",
		"Parcel Area Exclude Stacked Parcel (Y)",
		"Center Profile Zoning",
		"Center ID Number",
		"Comp Plan Area Name",
		"Comp Plan Type Name",
		"Comp Plan Type Code",
		"Shape__Area",
		"Shape__Length",
	}
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
			log.Fatalf("required column %s not found in CSV", target)
		}
	}

	fReader := &filteredReader{
		source:  reader,
		indices: indices,
	}

	// 4. Create Table
	if err := createTable(db, "parcel_geo", newHeader); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 5. Insert Rows
	count, err := insertParcels(db, "parcel_geo", newHeader, fReader)
	if err != nil {
		log.Fatalf("failed to insert parcel geo: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into data/parcels.db (parcel_geo)\n", count)
}

func loadCondos() {
	// 1. Open CSV and read header
	header, reader, csvFile, err := getCSVReader("data/condos.csv")
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

	// 3. Filter columns
	targetCols := []string{
		"Major",
		"Minor",
		"UnitType",
		"BldgNbr",
		"UnitNbr",
		"PcntOwnership",
		"BuildingNumber",
		"Fraction",
		"DirectionPrefix",
		"StreetName",
		"StreetType",
		"DirectionSuffix",
		"UnitDescr",
		"ZipCode",
	}
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
			log.Fatalf("required column %s not found in CSV", target)
		}
	}

	fReader := &filteredReader{
		source:  reader,
		indices: indices,
	}

	// 4. Create Table
	if err := createTable(db, "condos", newHeader); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 5. Insert Rows
	count, err := insertParcels(db, "condos", newHeader, fReader)
	if err != nil {
		log.Fatalf("failed to insert condos: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into data/parcels.db (condos)\n", count)
}

func loadResidentialBuildings() {
	// 1. Open CSV and read header
	header, reader, csvFile, err := getCSVReader("data/residential-buildings.csv")
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

	// 3. Filter columns
	targetCols := []string{
		"Major",
		"Minor",
		"BldgNbr",
		"NbrLivingUnits",
		"BuildingNumber",
		"Fraction",
		"DirectionPrefix",
		"StreetName",
		"StreetType",
		"DirectionSuffix",
		"ZipCode",
		"YrBuilt",
		"YrRenovated",
	}
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
			log.Fatalf("required column %s not found in CSV", target)
		}
	}

	fReader := &filteredReader{
		source:  reader,
		indices: indices,
	}

	// 4. Create Table
	if err := createTable(db, "residential_buildings", newHeader); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// 5. Insert Rows
	count, err := insertParcels(db, "residential_buildings", newHeader, fReader)
	if err != nil {
		log.Fatalf("failed to insert residential buildings: %v", err)
	}

	fmt.Printf("Successfully inserted %d records into data/parcels.db (residential_buildings)\n", count)
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
