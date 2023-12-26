package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Migration struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	cfg      Config
	db       *sql.DB
}

func (m *Migration) checkTable() bool {
	rows, tableCheck := m.db.Query(fmt.Sprintf("SELECT * FROM %s", m.cfg.migrationsTable))

	if tableCheck == nil {
		m.infoLog.Printf("Table %s found\n", m.cfg.migrationsTable)
		rows.Close()
		return true
	}
	m.infoLog.Printf("Table %s not found.", m.cfg.migrationsTable)
	return false
}

func (m *Migration) createTable() {
	stmt := fmt.Sprintf(`CREATE TABLE %s (
			 id INTEGER NOT NULL PRIMARY KEY,
			 applied DATETIME NOT NULL
			);`, m.cfg.migrationsTable)
	m.db.Exec(stmt)
}

func (m *Migration) getLastMigration() (int, error) {
	stmt := fmt.Sprintf("SELECT id FROM %s ORDER BY id DESC LIMIT 1", m.cfg.migrationsTable)

	var last sql.NullInt32

	err := m.db.QueryRow(stmt).Scan(&last)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		} else {
			return -1, err
		}
	}

	return int(last.Int32), nil
}

func (m *Migration) getMigrationFiles(last int) ([]MigrationFile, error) {
	var files []MigrationFile

	entries, err := os.ReadDir(m.cfg.migrationsFolder)
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

func (m *Migration) applyMigrations(files []MigrationFile) error {
	for _, entry := range files {
		m.infoLog.Printf("Executing migration %d", entry.id)

		filePath := filepath.Join(m.cfg.migrationsFolder, entry.File.Name())
		dat, err := os.ReadFile(filePath)
		check(err)

		document := strings.Split(string(dat), ";")
		for _, sentence := range document {
			if sentence != "" {
				fmt.Printf("Executing %s\n", sentence)
				_, err := m.db.Exec(sentence)
				check(err)
			}
		}
		stmt := fmt.Sprintf("INSERT INTO %s (id, applied) VALUES (?, UTC_TIMESTAMP())", m.cfg.migrationsTable)
		result, err := m.db.Exec(stmt, entry.id)
		check(err)
		_, err = result.LastInsertId()
		check(err)
	}
	return nil
}
