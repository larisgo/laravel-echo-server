package database

import (
	"database/sql"
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/utils"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDatabase struct {

	// SQLite client.
	sqlite *sql.DB
}

// Create a new cache instance.
func NewSQLiteDatabase(_options *options.Config) (DatabaseDriver, error) {
	db := &SQLiteDatabase{}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	sqlite_db := filepath.Clean(path.Join(cwd, _options.DatabaseConfig.Sqlite.DatabasePath))
	if path := filepath.Dir(sqlite_db); !utils.Exists(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}
	db.sqlite, err = sql.Open("sqlite3", sqlite_db)
	if err != nil {
		return nil, err
	}

	if _, err = db.sqlite.Exec(`CREATE TABLE IF NOT EXISTS key_value (key VARCHAR(255), value TEXT);CREATE UNIQUE INDEX IF NOT EXISTS key_index ON key_value (key);`); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *SQLiteDatabase) Close() error {
	return db.sqlite.Close()
}

// Retrieve data from redis.
func (db *SQLiteDatabase) Get(key string) ([]byte, error) {
	rows, err := db.sqlite.Query("SELECT value FROM key_value WHERE key = ? LIMIT 1", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var value []byte
	if rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return value, nil
}

// Store data to cache.
func (db *SQLiteDatabase) Set(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.sqlite.Exec("INSERT OR REPLACE INTO key_value (key, value) VALUES (?, ?)", key, data)
	return err
}
