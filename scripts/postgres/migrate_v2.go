package main

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	pqmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

// ! this is not in use yet?
func main() {
	// * create table if not exist profiles & bookings & (??)
	// * UP and DOWN:
	// UP: Copy all data from production db and create new tables/entities, while
	// the previous db still running in production.
	// DOWN: Start over from now database if data is fucked

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_CONNSTR"))
	if err != nil {
		panic(err)
	}
	driver, err := pqmigrate.WithInstance(db, &pqmigrate.Config{})
	m, err := migrate.NewWithDatabaseInstance("", "postgres", driver)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		panic(err)
	}
}

/**
 * Fill out this shit eventually
 */
