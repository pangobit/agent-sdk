# MCP Project Guidelines

This document outlines the conventions and commands for the MCP project.

## Project Overview

- **Purpose:** A simple, internal implementation of MCP for Pangobit, LLC.
- **License:** MIT
- **Primary Dependency:** Go Standard Library

## Code Style & Conventions

- **Paradigm:** Functional and data-driven concepts are favored. Avoid Object-Oriented Programming (OOP).
- **Naming:** Follow idiomatic Go naming conventions.
- **Project Structure:** Adhere to the standard Go project layout.
- **Error Handling:** Use `error` return values; avoid panics for recoverable errors.
- **Formatting:** Code must be formatted with `gofmt` or `goimports`.

## Development Commands

- **Build:** `go build ./...`
- **Test:** `go test ./...`
- **Run a single test:** `go test -run ^TestMyFunction$` (replace `TestMyFunction` with the test name)
- **Lint:** `golangci-lint run` (assuming `golangci-lint` is installed)
- **Tidy Dependencies:** `go mod tidy`
