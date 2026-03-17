package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql schema_mysql.sql
var schemaFS embed.FS

// Driver 数据库驱动类型
type Driver string

const (
	DriverSQLite Driver = "sqlite"
	DriverMySQL  Driver = "mysql"
)

// Open 根据 driver 打开数据库，自动初始化 schema
func Open(driver Driver, dsn string) (*sql.DB, error) {
	switch driver {
	case DriverSQLite:
		return OpenSQLite(dsn)
	case DriverMySQL:
		return OpenMySQL(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// OpenSQLite 打开或创建 SQLite 数据库，若不存在则初始化 schema
func OpenSQLite(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "gaojia.db"
	}
	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	// modernc.org/sqlite 为纯 Go 实现，支持 CGO_ENABLED=0 交叉编译
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, fmt.Errorf("exec schema: %w", err)
	}

	return db, nil
}

// OpenMySQL 打开 MySQL 连接，若表不存在则初始化 schema
func OpenMySQL(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn required")
	}
	// 确保 parseTime、multiStatements 以正确解析 DATETIME 并执行多语句 schema
	params := []string{}
	if !strings.Contains(dsn, "charset=") {
		params = append(params, "charset=utf8mb4")
	}
	if !strings.Contains(dsn, "parseTime=true") {
		params = append(params, "parseTime=true")
	}
	if !strings.Contains(dsn, "multiStatements=true") {
		params = append(params, "multiStatements=true")
	}
	if len(params) > 0 {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + strings.Join(params, "&")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	schema, err := schemaFS.ReadFile("schema_mysql.sql")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, fmt.Errorf("exec schema: %w", err)
	}

	return db, nil
}
