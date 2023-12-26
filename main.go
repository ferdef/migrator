package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const MIGRATIONS_TABLE = "migrations"
const MIGRATIONS_PATH = "./internal/db/migrations/"

type MigrationFile struct {
	id   int
	File fs.DirEntry
}

type Config struct {
	user             string
	pass             string
	ip               string
	db               string
	dsn              string
	createTable      bool
	migrationsTable  string
	migrationsFolder string
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	config := parseConfig()

	infoLog.Print("Connecting to DB...")

	db, err := openDB(config.dsn)
	check(err)

	migration := Migration{
		infoLog:  infoLog,
		errorLog: errorLog,
		cfg:      config,
		db:       db,
	}

	defer db.Close()

	infoLog.Println("Connected")

	if !migration.checkTable() {
		if config.createTable {
			migration.createTable()
		} else {
			errorLog.Fatalf("Table %s doesn't exist and we're not creating it. Exiting...", config.migrationsTable)
		}

	}

	// last, err := getLastMigration(db, )
	// check(err)
	// infoLog.Printf("Last migration applied %d\n", last)

	// files, err := getMigrationFiles(MIGRATIONS_PATH, last)
	// check(err)

	// fmt.Printf("Pending migrations: \n%v", files)
	// err = applyMigrations(db, files)
	// check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseConfig() Config {
	var cfg Config

	flag.StringVar(&cfg.user, "user", "", "MySQL User")
	flag.StringVar(&cfg.pass, "pass", "", "MySQL Password")
	flag.StringVar(&cfg.ip, "ip", "", "MySQL IP")
	flag.StringVar(&cfg.db, "db", "", "MySQL Database Name")
	flag.StringVar(&cfg.migrationsFolder, "folder", MIGRATIONS_PATH, "Migrations Folder")
	flag.StringVar(&cfg.migrationsTable, "table", MIGRATIONS_TABLE, "Migrations Table")
	flag.BoolVar(&cfg.createTable, "c", false, "Create migrations table if not exists?")
	flag.Parse()
	checkEmpty(cfg.user)
	checkEmpty(cfg.pass)
	checkEmpty(cfg.ip)
	checkEmpty(cfg.db)
	checkEmpty(cfg.migrationsFolder)
	checkEmpty(cfg.migrationsTable)

	cfg.dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.user, cfg.pass, cfg.ip, cfg.db)

	return cfg
}

func checkEmpty(test string) {
	if test == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	check(err)

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
