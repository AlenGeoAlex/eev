package config

func (d DBType) IsValid() bool {
	switch d {
	case DBTypeSqlite:
		return true
	default:
		return false
	}
}

func (d DBType) MigrationPath() string {
	switch d {
	case DBTypeSqlite:
		return "file://migrations/sqlite/"
	default:
		panic("unsupported database type")
	}
}
