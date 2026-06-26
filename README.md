# Seattle Scraper

Seattle Scraper is a research project to explore mapping real estate parcels
with their legal owners. The idea is to build a database capable of answering
questions such as:

- How many homes in Seattle are owned by corporations?
- Who owns the building at 123 Somewhere Ave W?
- How many people own multiple properties in Seattle?

A secondary goal of this project is to explore using local LLMs to improve
software engineering.

## Data sources

- https://info.kingcounty.gov/assessor/DataDownload/default.aspx

## Approach

From the book of Gemini:

To map out who owns what, you need to tie spatial geographic data (the land
parcels) to the tax assessor's data (the owner records). The King County
Department of Assessments provides these database extracts for free.

The King County Assessor Data Download Portal is where you grab the raw
relational tables. You can download text/CSV extracts of the entire county
database. The critical file you want is the parcel_extr (Parcel Record Assessor
extract table), which contains fields for ownership name, taxpayer name, mailing
address, and property characteristics.

To get the actual geographic boundaries of the properties, you can download the
King County Tax Parcel Polygons via the King County GIS open data portal in
GeoJSON, Shapefile, or File Geodatabase formats.

Both datasets use a unique 10-digit PIN (Parcel Identification Number), which is
split into a 6-digit MAJOR code (the neighborhood/block) and a 4-digit MINOR
code (the specific lot). You will use MAJOR + MINOR as your primary keys to join
the geographic shapes to the text-based ownership data in your local database.

## Prompts

- Write a Go CLI named "parcel-reader". It should do nothing more than print
  "hello" and then exit cleanly.
