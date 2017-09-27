package main

import (
	"database/sql"
	"fmt"
	"time"

	"analysis-tools/lib"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var (
	oriDb *sql.DB
	endDb *sql.DB
)

type DailyStat struct {
	NumUsers     int
	NumThreads   int
	NumPosts     int
	NumEnThreads int
	NumEnPosts   int
}

func init() {
	var err error
	cfg, err := lib.LoadDefaultConfig()
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "load config failed"))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.BbsDb.User, cfg.BbsDb.Password, cfg.BbsDb.Host, cfg.BbsDb.PortStr(), cfg.BbsDb.Database)
	oriDb, err = sql.Open("mysql", dsn)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "connect db failed"))
	}

	endDb, err = lib.ConnectDB(cfg)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "connect db failed"))
	}
}

func main() {
	const sqlQuery = `
	SELECT count(*) FROM pre_common_member UNION
	SELECT sum(threads) FROM pre_forum_forum UNION
	SELECT sum(posts) FROM pre_forum_forum UNION
	SELECT threads FROM pre_forum_forum WHERE fid = 70 UNION
	SELECT posts FROM pre_forum_forum WHERE fid = 70
	`
	rows, err := oriDb.Query(sqlQuery)
	if err != nil {
		lib.Logger.Fatal(err)
	}
	d := DailyStat{}
	rows.Next()
	rows.Scan(&d.NumUsers)
	rows.Next()
	rows.Scan(&d.NumThreads)
	rows.Next()
	rows.Scan(&d.NumPosts)
	rows.Next()
	rows.Scan(&d.NumEnThreads)
	rows.Next()
	rows.Scan(&d.NumEnPosts)

	ntime := time.Now()
	_, err = endDb.Exec(`
	INSERT INTO bbs_daily_stat (num_user, num_zh_thread, num_zh_post, num_en_thread, num_en_post, time)
	VALUES($1, $2, $3, $4, $5, $6)
	`, d.NumUsers, d.NumThreads-d.NumEnThreads, d.NumPosts-d.NumEnPosts, d.NumEnThreads, d.NumEnPosts, ntime)
	if err != nil {
		lib.Logger.Fatal(err)
	}
}
