package main

import (
	"flag"
	"fmt"
	"os"
)

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
