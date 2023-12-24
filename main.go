package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

var MIGRATIONS_PATH = "./internal/db/migrations/"

type MigrationFile struct {
	id   int
	File fs.DirEntry
}

type Config struct {
	user string
	pass string
	ip string
	db string
	dsn string
}

func main() {
	config := parseConfig()

	infoLog.Print("Connecting to DB...")
	
	db, err := openDB(config.dsn)
	check(err)

	infoLog.Println("Connected")

	defer db.Close()

	// last, err := getLastMigration(db)
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
	flag.Parse()
	checkEmpty("user", cfg.user)
	checkEmpty("pass", cfg.pass)
	checkEmpty("ip", cfg.ip)
	checkEmpty("db", cfg.db)

	cfg.dsn =  fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.user, cfg.pass, cfg.ip, cfg.db)

	return cfg
}


func checkEmpty(name string, test string) {
	if (test == "") {
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

func getLastMigration(db *sql.DB) (int, error) {
	stmt := "SELECT id FROM migrations ORDER BY id DESC LIMIT 1"

	var last sql.NullInt32

	err := db.QueryRow(stmt).Scan(&last)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		} else {
			return -1, err
		}

	}

	return int(last.Int32), nil
}

func getMigrationFiles(dir string, last int) ([]MigrationFile, error) {
	var files []MigrationFile

	entries, err := os.ReadDir(dir)
	check(err)

	for _, e := range entries {
		if e.Type().IsRegular() {
			strId := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			id, err := strconv.Atoi(strId)
			check(err)

			if id > last {
				file := MigrationFile{
					id:   id,
					File: e,
				}
				files = append(files, file)
			}
		}
	}

	return files, nil
}

func applyMigrations(db *sql.DB, files []MigrationFile) error {
	for _, entry := range files {
		fmt.Printf("Executing migration %d\n", entry.id)

		filePath := filepath.Join(MIGRATIONS_PATH, entry.File.Name())
		dat, err := os.ReadFile(filePath)
		check(err)

		document := strings.Split(string(dat), ";")
		for _, sentence := range document {
			if sentence != "" {
				fmt.Printf("Executing %s\n", sentence)
				_, err := db.Exec(sentence)
				check(err)
			}
		}
		stmt := "INSERT INTO migrations (id, applied) VALUES (?, UTC_TIMESTAMP())"
		result, err := db.Exec(stmt, entry.id)
		check(err)
		_, err = result.LastInsertId()
		check(err)
	}
	return nil
}
