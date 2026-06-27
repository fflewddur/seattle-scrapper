package main

import (
	"log"

	_ "modernc.org/sqlite"
)

type TableConfig struct {
	CSVPath    string
	TableName  string
	TargetCols []string
}

func main() {
	configs := []TableConfig{
		{
			CSVPath:   "data/parcels.csv",
			TableName: "parcels",
			TargetCols: []string{
				"Major", "Minor", "PropName", "PropType", "PlatName", "PlatLot", "PlatBlock",
			},
		},
		{
			CSVPath:   "data/parcel_geo.csv",
			TableName: "parcel_geo",
			TargetCols: []string{
				"Major", "Minor", "Site Address (per KCA)", "Property Name (per KCA)",
				"Property Type (per KCA)", "Land Use Code (per KCA)", "Detailed Existing Land Use (per KCA)",
				"Site Zip Code (per KCA)", "Building Description (per KCA)", "Ownership Type",
				"Public Ownership Category", "Parcel Area Exclude Stacked Parcel (Y)",
				"Center Profile Zoning", "Center ID Number", "Comp Plan Area Name",
				"Comp Plan Type Name", "Comp Plan Type Code", "Shape__Area", "Shape__Length",
			},
		},
		{
			CSVPath:   "data/condos.csv",
			TableName: "condos",
			TargetCols: []string{
				"Major", "Minor", "UnitType", "BldgNbr", "UnitNbr", "PcntOwnership",
				"BuildingNumber", "Fraction", "DirectionPrefix", "StreetName",
				"StreetType", "DirectionSuffix", "UnitDescr", "ZipCode",
			},
		},
		{
			CSVPath:   "data/residential-buildings.csv",
			TableName: "residential_buildings",
			TargetCols: []string{
				"Major", "Minor", "BldgNbr", "NbrLivingUnits", "BuildingNumber",
				"Fraction", "DirectionPrefix", "StreetName", "StreetType",
				"DirectionSuffix", "ZipCode", "YrBuilt", "YrRenovated",
			},
		},
		{
			CSVPath:    "data/apartments.csv",
			TableName:  "apartments",
			TargetCols: []string{"Major", "Minor", "ComplexDescr", "Address"},
		},
	}

	for _, cfg := range configs {
		if err := loadTable(cfg.CSVPath, cfg.TableName, cfg.TargetCols); err != nil {
			log.Printf("failed to load %s: %v", cfg.TableName, err)
		}
	}
}
