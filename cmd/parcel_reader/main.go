package main

import (
	_ "modernc.org/sqlite"
)

func main() {
	loadParcels("data/parcels.csv", "parcels", []string{"Major", "Minor", "PropName", "PropType", "PlatName", "PlatLot", "PlatBlock"})
	loadParcelGeo("data/parcel_geo.csv", "parcel_geo", []string{
		"Major", "Minor", "Site Address (per KCA)", "Property Name (per KCA)",
		"Property Type (per KCA)", "Land Use Code (per KCA)", "Detailed Existing Land Use (per KCA)",
		"Site Zip Code (per KCA)", "Building Description (per KCA)", "Ownership Type",
		"Public Ownership Category", "Parcel Area Exclude Stacked Parcel (Y)",
		"Center Profile Zoning", "Center ID Number", "Comp Plan Area Name",
		"Comp Plan Type Name", "Comp Plan Type Code", "Shape__Area", "Shape__Length",
	})
	loadCondos("data/condos.csv", "condos", []string{
		"Major", "Minor", "UnitType", "BldgNbr", "UnitNbr", "PcntOwnership",
		"BuildingNumber", "Fraction", "DirectionPrefix", "StreetName",
		"StreetType", "DirectionSuffix", "UnitDescr", "ZipCode",
	})
	loadResidentialBuildings("data/residential-buildings.csv", "residential_buildings", []string{
		"Major", "Minor", "BldgNbr", "NbrLivingUnits", "BuildingNumber",
		"Fraction", "DirectionPrefix", "StreetName", "StreetType",
		"DirectionSuffix", "ZipCode", "YrBuilt", "YrRenovated",
	})
	loadApartments("data/apartments.csv", "apartments", []string{"Major", "Minor", "ComplexDescr", "Address"})
}
