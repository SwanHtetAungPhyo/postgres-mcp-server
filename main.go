package main

import (
	"fmt"
	"os"
	"strings"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Set your PostgreSQL connection string here
// Example: "host=localhost user=postgres password=mysecretpassword dbname=mydb port=5432 sslmode=disable"
const dbConn = ""

// SQLArgs represents the structure for SQL query arguments
type SQLArgs struct {
	Query string `json:"query" jsonschema:"required,description=Raw SQL query to execute"`
}

func main() {
	fmt.Fprintln(os.Stderr, "Starting MCP server...")

	done := make(chan struct{})

	server := mcp.NewServer(stdio.NewStdioServerTransport())

	db, err := gorm.Open(postgres.Open(dbConn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB connection failed: %v\n", err)
		os.Exit(1)
	}

	// Register tool for executing SELECT queries
	server.RegisterTool("execute_query", "Execute raw SQL SELECT", func(args SQLArgs) (*mcp.ToolResponse, error) {
		if !hasPrefix(args.Query, "SELECT") {
			return nil, fmt.Errorf("only SELECT queries are allowed")
		}

		rows, err := db.Raw(args.Query).Rows()
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		var output string

		for rows.Next() {
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			err := db.ScanRows(rows, &values)
			if err != nil {
				return nil, err
			}

			for i, col := range columns {
				output += fmt.Sprintf("%s: %v\t", col, values[i])
			}
			output += "\n"
		}

		return mcp.NewToolResponse(mcp.NewTextContent(output)), nil
	})

	// Register tool for DDL queries (CREATE, DROP, ALTER)
	server.RegisterTool("ddl_query", "Run a DDL query (CREATE, DROP, ALTER)", func(args SQLArgs) (*mcp.ToolResponse, error) {
		ddlPrefixes := []string{"CREATE", "DROP", "ALTER"}
		if !hasAnyPrefix(args.Query, ddlPrefixes...) {
			return nil, fmt.Errorf("only CREATE, DROP, or ALTER queries are allowed")
		}
		return runExec(db, args.Query)
	})

	// Register tool for DML queries (INSERT, UPDATE, DELETE)
	server.RegisterTool("modify_query", "Run a DML query (INSERT, UPDATE, DELETE)", func(args SQLArgs) (*mcp.ToolResponse, error) {
		dmlPrefixes := []string{"INSERT", "UPDATE", "DELETE"}
		if !hasAnyPrefix(args.Query, dmlPrefixes...) {
			return nil, fmt.Errorf("only INSERT, UPDATE, or DELETE queries are allowed")
		}
		return runExec(db, args.Query)
	})

	fmt.Fprintln(os.Stderr, "Server initialized, starting Serve()...")

	// Start the server in a goroutine
	go func() {
		if err := server.Serve(); err != nil {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Fprintln(os.Stderr, "Server running, waiting for done signal...")
	<-done
	fmt.Fprintln(os.Stderr, "Server shutting down...")
}

// hasPrefix checks if a query starts with a specific prefix (case-insensitive)
func hasPrefix(query, prefix string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), prefix)
}

// hasAnyPrefix checks if a query starts with any of the specified prefixes (case-insensitive)
func hasAnyPrefix(query string, prefixes ...string) bool {
	query = strings.ToUpper(strings.TrimSpace(query))
	for _, prefix := range prefixes {
		if strings.HasPrefix(query, prefix) {
			return true
		}
	}
	return false
}

// runExec executes a query and returns a success response
func runExec(db *gorm.DB, query string) (*mcp.ToolResponse, error) {
	if err := db.Exec(query).Error; err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Query executed successfully.")), nil
}
