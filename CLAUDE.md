# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Palette is a Go library for working with collections of colors. It provides a unified interface for importing and exporting color palettes in various formats including Adobe Color Book (.acb), Adobe Color Swatch (.aco), CSV, and JSON.

## Architecture

The project is organized around a composable architecture with color types, palette collections, and pluggable import/export system:

- `color/` - Core color types and interfaces (RGB, CMYK, LAB, HSB) with conversion methods
- `palette/` - Palette collection type for managing groups of colors
- `io/` - Import/export framework with pluggable format support
- `io/colorbook/` - Adobe Color Book format (.acb files) importer/exporter
- `io/colorswatch/` - Adobe Color Swatch format (.aco files) importer/exporter
- `io/csv/` - CSV format importer/exporter with multiple color representations
- `io/json/` - JSON format importer/exporter with flexible schema support
- `adobe/` - Legacy Adobe-specific parsing (used by format importers)

The codebase uses Go's `stringer` tool to generate string representations for enums (BookID, ColorType, ColorSpace).

## Development Commands

- **Build**: `go build ./...`
- **Generate code**: `go generate ./...` (generates string methods for enums)
- **Test**: `go test ./...`
- **Format**: `go fmt ./...`

## Code Generation

The project uses `go generate` with the `stringer` tool to generate string methods for enum types. Generated files follow the pattern `*_string.go` and should not be manually edited.

## Binary Format Parsing

The colorbook package implements Adobe's Color Book specification using binary.Read with BigEndian byte order. String parsing involves UTF-16 to UTF-8 conversion for Adobe's text encoding format.