package config

import "log"

func validate(cfg *DBConfig) {

	if !cfg.DbType.IsValid() {
		log.Fatalf("Invalid DB_TYPE: %s", cfg.DbType)
	}

	if cfg.DbType == DBTypeSqlite && cfg.SqlitePath == "" {
		log.Fatal("DB_SQLITE_PATH required for sqlite")
	}

}
