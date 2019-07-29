package database

type DatabaseDriver interface {
	/**
	 * Get a value from the database.
	 */
	Get(string) (interface{}, error)

	/**
	 * Set a value to the database.
	 */
	Set(string, interface{}) error
}
