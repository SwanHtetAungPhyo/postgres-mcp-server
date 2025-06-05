# PostgreSQL MCP Server

[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![MCP](https://img.shields.io/badge/MCP-Compatible-blue?style=flat)](https://modelcontextprotocol.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/postgresql-mcp-server)](https://goreportcard.com/report/github.com/yourusername/postgresql-mcp-server)
# Author: Swan Htet Aung Phyo 
# Computer Science Student At AGH (Backend Developer)
> A secure and efficient MCP server that lets AI assistants interact with PostgreSQL databases through a clean, standardized interface.

Ever wanted to give your AI assistant the ability to query your database without worrying about security? This MCP server is exactly what you need. It provides a safe way for AI models to interact with PostgreSQL databases while keeping your data secure.

## What's This All About?

The Model Context Protocol (MCP) is becoming the standard way for AI assistants to connect with external data sources. This server implements MCP specifically for PostgreSQL databases, giving you three main capabilities:

- **Query your data** - Let AI assistants run SELECT queries to find information
- **Manage your schema** - Create, modify, or drop database structures when needed  
- **Update your data** - Insert, update, or delete records with proper validation

The best part? Everything is validated and restricted to prevent dangerous operations. No accidental `DROP DATABASE` commands here!

## Getting Started

### What You'll Need

- Go 1.19 or later installed on your machine
- A PostgreSQL database (version 13+ recommended)
- About 5 minutes to get everything running

### Installation

First, let's create your project:

```bash
mkdir postgresql-mcp-server
cd postgresql-mcp-server
go mod init postgresql-mcp-server
```

Now grab the dependencies:

```bash
go get github.com/metoro-io/mcp-golang
go get github.com/metoro-io/mcp-golang/transport/stdio
go get gorm.io/driver/postgres
go get gorm.io/gorm
```

Copy the main server code into `main.go`, then update your database connection:

```go
const dbConn = "host=localhost user=postgres password=yourpassword dbname=yourdb port=5432 sslmode=disable"
```

Build and run:

```bash
go build -o postgresql-mcp-server
./postgresql-mcp-server
```

That's it! Your server should be running and ready to accept connections.

## How to Use It

The server provides three main tools, each designed for specific types of database operations:

### Reading Data (`execute_query`)

This is probably what you'll use most often. It lets you run SELECT queries to retrieve information:

```sql
-- Find all active users
SELECT name, email FROM users WHERE active = true;

-- Get sales summary for the last month
SELECT 
    product_name,
    SUM(quantity) as total_sold,
    AVG(price) as avg_price
FROM sales 
WHERE created_at >= '2024-01-01'
GROUP BY product_name
ORDER BY total_sold DESC;
```

The results come back nicely formatted, making it easy for AI assistants to understand and work with your data.

### Managing Database Structure (`ddl_query`)

Need to create tables or modify your schema? This tool handles DDL operations:

```sql
-- Create a new table
CREATE TABLE blog_posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id INTEGER REFERENCES users(id),
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add an index for better performance
CREATE INDEX idx_posts_author ON blog_posts(author_id);

-- Modify existing table
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
```

### Modifying Data (`modify_query`)

When you need to insert, update, or delete data:

```sql
-- Add new records
INSERT INTO users (name, email, active) 
VALUES ('Alice Johnson', 'alice@example.com', true);

-- Update existing data
UPDATE products 
SET price = price * 0.9 
WHERE category = 'electronics' AND stock > 100;

-- Clean up old data
DELETE FROM logs WHERE created_at < NOW() - INTERVAL '90 days';
```

## Security Features

Security was a major consideration when building this server. Here's how we keep your data safe:

**Query Validation**: Every query is checked before execution. Only specific types of SQL statements are allowed for each tool. You can't accidentally run a DROP TABLE command through the query tool.

**No Dangerous Operations**: Commands like TRUNCATE, GRANT, or other administrative functions are blocked entirely.

**Connection Security**: Uses standard PostgreSQL connection security including SSL support.

**Input Sanitization**: All queries go through GORM's built-in protections against SQL injection.

## Configuration Options

### Database Connection

The connection string format is standard PostgreSQL:

```
host=localhost user=myuser password=mypass dbname=mydb port=5432 sslmode=require
```

For production environments, consider these security settings:

```go
// Production example
const dbConn = "host=db.example.com user=mcp_user password=secure_password dbname=production_db port=5432 sslmode=require"
```

### Environment Variables

For better security, you might want to use environment variables instead of hardcoding credentials:

```go
import "os"

func getDBConnection() string {
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")
    
    return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require", 
        host, user, password, dbname, port)
}
```

## Connecting to MCP Clients

### Claude Desktop

Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "postgresql": {
      "command": "/path/to/your/postgresql-mcp-server",
      "args": []
    }
  }
}
```

### Other MCP Clients

Most MCP clients use a similar configuration format:

```json
{
  "servers": [
    {
      "name": "postgresql-db",
      "transport": "stdio", 
      "command": ["/path/to/postgresql-mcp-server"]
    }
  ]
}
```

## Real-World Examples

Here are some practical ways you might use this server:

### Business Intelligence Queries
```sql
-- Monthly revenue trend
SELECT 
    DATE_TRUNC('month', order_date) as month,
    SUM(total_amount) as revenue,
    COUNT(*) as order_count
FROM orders 
WHERE order_date >= '2024-01-01'
GROUP BY month
ORDER BY month;
```

### User Analytics
```sql
-- Active users by registration date
SELECT 
    DATE(created_at) as signup_date,
    COUNT(*) as new_users,
    SUM(COUNT(*)) OVER (ORDER BY DATE(created_at)) as cumulative_users
FROM users 
WHERE created_at >= '2024-01-01'
GROUP BY signup_date
ORDER BY signup_date;
```

### Inventory Management
```sql
-- Low stock alert
SELECT 
    product_name,
    current_stock,
    min_stock_level,
    (min_stock_level - current_stock) as shortage
FROM inventory 
WHERE current_stock < min_stock_level
ORDER BY shortage DESC;
```

## Common Issues and Solutions

**Connection Problems**: If you can't connect to your database, double-check your connection string. Make sure PostgreSQL is running and accepting connections on the specified port.

**Permission Errors**: The database user needs appropriate permissions for the operations you want to perform. For read-only access, SELECT permissions are enough. For full functionality, you'll need CREATE, INSERT, UPDATE, and DELETE permissions.

**Query Blocked**: If your query gets rejected, check that you're using the right tool. SELECT queries go to `execute_query`, while CREATE/ALTER/DROP go to `ddl_query`.

**Performance Issues**: For large datasets, consider adding LIMIT clauses to your queries. The server doesn't automatically limit result sets, so a query that returns millions of rows might be slow.

## Performance Tips

- Use indexes on columns you query frequently
- Add LIMIT clauses for large datasets
- Consider using EXPLAIN to optimize complex queries
- Keep your PostgreSQL statistics up to date with ANALYZE

## Contributing

Found a bug or want to add a feature? Contributions are welcome! Here's how:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/cool-new-feature`)
3. Make your changes
4. Test thoroughly
5. Commit your changes (`git commit -am 'Add cool new feature'`)
6. Push to the branch (`git push origin feature/cool-new-feature`)
7. Create a Pull Request

Please make sure your code follows Go conventions and includes appropriate error handling.

## License

This project is released under the MIT License. See the LICENSE file for details.

## Tags

`#mcp` `#postgresql` `#golang` `#ai` `#database` `#model-context-protocol` `#sql` `#gorm` `#claude` `#assistant` `#server` `#api` `#data` `#query` `#secure` `#open-source`

---

**Questions?** Open an issue on GitHub or start a discussion. We're here to help!