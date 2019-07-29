package database

import (
	"github.com/larisgo/laravel-echo-server/log"
	"github.com/larisgo/laravel-echo-server/options"
)

type Database struct {
	/**
	 * Database driver.
	 */
	driver DatabaseDriver
}

/**
 * Create a new database instance.
 */
func NewDatabase(Options options.Config) DatabaseDriver {
	this := &Database{}
	if Options.Database == "redis" {
		this.driver = NewRedisDatabase(Options)
	} else if Options.Database == "sqlite" {
		this.driver = NewSQLiteDatabase(Options)
	} else {
		log.Fatal("Database driver not set.")
	}
	return DatabaseDriver(this)
}

/**
 * Get a value from the database.
 */
func (this *Database) Get(key string) (interface{}, error) {
	return this.driver.Get(key)
}

/**
 * Set a value to the database.
 */
func (this *Database) Set(key string, value interface{}) error {
	return this.driver.Set(key, value)
}
