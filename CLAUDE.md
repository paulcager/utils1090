# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based ADS-B (Mode S) aircraft tracking utility suite for processing and visualizing data from Mode S/ADS-B receivers. The project processes SBS-1 BaseStation format messages commonly used by dump1090 and similar receivers.

**Package name:** `save1090` (at root level)

## Architecture

### Core Components

**messages.go** (root package `save1090`)
- Defines the `SBSMessage` struct representing parsed SBS-1 BaseStation format messages
- SBS message format reference: http://woodair.net/sbs/article/barebones42_socket_data.htm
- Contains decoder functions for parsing CSV-formatted aircraft position messages
- Key fields: HexIdent (Mode S code), Callsign, Altitude, Groundspeed, Track, LatLon, VerticalRate, Squawk
- Note: Altitude is Mode C Flight Level (not AMSL), Groundspeed is not airspeed

**cmd/cover1090/** - Web-based coverage visualization tool
- HTTP server on port 8080
- Reads lat/lon pairs from `/home/paul/locs.csv` (hardcoded path)
- Generates PNG coverage maps showing receiver range
- Serves Google Maps-based web interface via templates
- Uses `github.com/golang/geo/r2` for geographic calculations
- Template system reads from `TEMPLATES` env var (defaults to "templates" directory)

### Data Pipeline Scripts

**cmd/cover1090/select_locs** - AWK script to extract position data from SBS logs
- Filters SBS messages with valid lat/lon (field 15 not empty)
- Outputs: timestamp, hexcode, lat, lon

**cmd/cover1090/load_db** - Database initialization and data loading
- Creates SQLite database with `locs` table (timestamp, hexcode, lat, lon)
- Processes compressed SBS logs (`~/192.168.0.15:30003-*.log*`)
- Uses `select_locs` to filter and import into SQLite

## Module Setup

**Module name:** `github.com/paulcager/utils1090`
**Go version:** 1.24.2

Dependencies (managed in go.mod):
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO-enabled, requires gcc)
- `github.com/golang/geo` - Geographic calculations

To update dependencies:
```bash
go mod tidy
```

## Building and Running

### Build cover1090 server

```bash
cd cmd/cover1090
go build -o cover1090
```

### Run cover1090 server

Required environment variables:
- `APIKey` - Google Maps API key (required for map display)
- `ZOOM` - Initial map zoom level (default: 8)
- `LAT` - Initial map center latitude (default: 53)
- `LON` - Initial map center longitude (default: -2.25)
- `TEMPLATES` - Template directory path (default: "templates")

```bash
APIKey=your_google_maps_api_key ./cover1090
```

Access the web interface at http://localhost:8080/

### Data Processing

```bash
cd cmd/cover1090
./load_db  # Creates locs.db and imports position data from SBS logs
```

## Code Conventions

- SBS message timestamps: Uses format "2006/01/02T15:04:05.000"
- Boolean flags in SBS: "-1" represents true
- Coordinate system: Standard lat/lon (WGS84)
- Image rendering: Y-axis is latitude, X-axis is longitude
- Bounds for visualization hardcoded to UK region: lat 52-54, lon -4 to 0
