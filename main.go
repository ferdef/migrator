package main

func main() {

	config := parseConfig()

	migration := Migration{
		cfg: config,
	}

	migration.migrate()
}
