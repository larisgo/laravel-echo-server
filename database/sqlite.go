package database

import (
	"database/sql"
	"encoding/json"
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
	"github.com/larisgo/laravel-echo-server/utils"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"path/filepath"
)

type SQLiteDatabase struct {
	/**
	 * SQLite client.
	 */
	sqlite *sql.DB
}

/**
 * Create a new cache instance.
 */
func NewSQLiteDatabase(Options options.Config) DatabaseDriver {
	this := &SQLiteDatabase{}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	sqlite_db := filepath.Clean(path.Join(cwd, Options.DatabaseConfig.Sqlite.DatabasePath))
	if path := filepath.Dir(sqlite_db); !utils.Exists(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatal(err)
		}
	}
	this.sqlite, err = sql.Open("sqlite3", sqlite_db)
	if err != nil {
		log.Fatal(err)
	}
	// defer this.sqlite.Close()
	if _, err = this.sqlite.Exec(`CREATE TABLE IF NOT EXISTS key_value (key VARCHAR(255), value TEXT);CREATE UNIQUE INDEX IF NOT EXISTS key_index ON key_value (key);`); err != nil {
		log.Fatal(err)
	}
	return DatabaseDriver(this)
}

/**
 * Retrieve data from redis.
 */
func (this *SQLiteDatabase) Get(key string) ([]byte, error) {
	rows, err := this.sqlite.Query("SELECT value FROM key_value WHERE key = ? LIMIT 1", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var value string
	for rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return []byte(value), nil
}

/**
 * Store data to cache.
 */
func (this *SQLiteDatabase) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = this.sqlite.Exec("INSERT OR REPLACE INTO key_value (key, value) VALUES (?, ?)", key, data)
	return err
}
