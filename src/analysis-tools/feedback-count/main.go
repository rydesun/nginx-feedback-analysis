package main

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"analysis-tools/lib"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var (
	cfg        *lib.Config
	feedbackDb *sql.DB
)

type Count struct {
	View   LangCount `json:"view"`
	Closed LangCount `json:"closed"`
}

type LangCount struct {
	ZH int `db:"zh-cn" json:"zh-cn"`
	EN int `db:"en" json:"en"`
}

func getLastDate(db *sql.DB) time.Time {
	var date time.Time
	db.QueryRow("SELECT date FROM feedback_stat ORDER BY date DESC LIMIT 1").Scan(&date)
	return date
}

func insertLog(db *sql.DB, date time.Time, count Count) error {
	jsonData, err := json.Marshal(count)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO feedback_stat (date, count) VALUES($1, $2)", date, jsonData)
	return err
}

func parseHtml(page string, filePattern string) (dateUrlPairs [][2]string) {
	r := regexp.MustCompile("<a href=\"(.*)\">(.*)</a>")
	res := r.FindAllStringSubmatch(page, -1)
	for _, elems := range res {
		r := regexp.MustCompile(filePattern)
		res := r.FindStringSubmatch(elems[2])
		if len(res) > 1 {
			date := res[1]
			url := elems[1]
			dateUrlPairs = append(dateUrlPairs, [2]string{date, url})

		}
	}
	return
}

func parseLog(feedbackDb *sql.DB, fr io.Reader) Count {
	gzr, err := gzip.NewReader(fr)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "new gzip reader failed"))
	}
	defer gzr.Close()

	count := Count{}
	reader := gzr
	var ids []string

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		r := regexp.MustCompile("\"[^\"]*\"")
		reqStr := strings.Trim(r.FindString(scanner.Text()), `"`)
		req := strings.Fields(reqStr)
		if len(req) != 3 {
			continue
		}
		method := req[0]
		reqPath := req[1]
		u, err := url.Parse(reqPath)
		if err != nil {
			continue
		}
		if method == "GET" && u.Path == "/feedback" {
			language := u.Query()["language"]
			if len(language) == 0 {
				continue
			}
			switch language[0] {
			case "zh-cn":
				count.View.ZH++
			case "en":
				count.View.EN++
			}
		}
		if method == "GET" && u.Path == "/feedback/close" {
			idSeg := u.Query()["id"]
			if len(idSeg) != 0 {
				if _, err = strconv.Atoi(idSeg[0]); err != nil {
					lib.Logger.Printf("broken close id: %s", idSeg[0])
					continue
				}
				ids = append(ids, idSeg[0])
			}
		}
	}
	if len(ids) != 0 {
		var language string
		rows, err := feedbackDb.Query("SELECT language FROM question WHERE id in (" + strings.Join(ids, ", ") + ")")
		if err != nil {
			lib.Logger.Fatalln("failed to query", err, ids)
		}
		for rows.Next() {
			rows.Scan(&language)
			switch language {
			case "zh_cn":
				count.Closed.ZH++
			case "en":
				count.Closed.EN++
			}
		}
	}
	return count
}

func init() {
	var err error
	cfg, err = lib.LoadDefaultConfig()
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "load config failed"))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.LogWatch.User, cfg.LogWatch.Password, cfg.LogWatch.Host, cfg.LogWatch.PortStr(), cfg.LogWatch.Database)
	feedbackDb, err = sql.Open("mysql", dsn)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "connect db failed"))
	}
}

func main() {
	db, err := lib.ConnectDB(cfg)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "connect db failed"))
	}

	resp, err := http.Get(cfg.LogWatch.UrlBase)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "get log page failed"))
	}
	if resp.StatusCode != http.StatusOK {
		lib.Logger.Fatal("response status: ", resp.Status)
	}

	rawPage, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "get log page failed"))
	}

	beginDate := getLastDate(db)
	dateUrlPairs := parseHtml(string(rawPage), cfg.LogWatch.FilePattern)
	for _, pair := range dateUrlPairs {
		dateStr := pair[0]
		url := pair[1]
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			lib.Logger.Println(err)
			continue
		}
		if date.After(beginDate) {
			resp, err := http.Get(fmt.Sprintf("%s/%s", cfg.LogWatch.UrlBase, url))
			if err != nil {
				lib.Logger.Println(err)
			}
			count := parseLog(feedbackDb, resp.Body)
			err = insertLog(db, date, count)
			if err != nil {
				lib.Logger.Println(err)
			}
		}
	}
}
