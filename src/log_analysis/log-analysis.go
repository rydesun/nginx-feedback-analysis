package log_analysis

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

var Logger *log.Logger

func init() {
	prefix := "[" + filepath.Base(os.Args[0]) + "] "
	Logger = log.New(os.Stdout, prefix, log.Lshortfile)
}

func ConnectDB(cfg *Config) (*sql.DB, error) {
	dbCfg := cfg.Db
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		dbCfg.Host, dbCfg.Port, dbCfg.Database, dbCfg.User, dbCfg.Password)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
