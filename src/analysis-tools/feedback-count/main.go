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
	"strings"
	"time"

	"analysis-tools/lib"

	"github.com/pkg/errors"
)

type LangCount struct {
	ZH int `db:"zh-cn" json:"zh-cn"`
	EN int `db:"en" json:"en"`
}

func getLastDate(db *sql.DB) time.Time {
	var date time.Time
	db.QueryRow("SELECT date FROM feedback_stat ORDER BY date DESC LIMIT 1").Scan(&date)
	return date
}

func insertLog(db *sql.DB, date time.Time, count LangCount) error {
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

func parseLog(fr io.Reader) LangCount {
	gzr, err := gzip.NewReader(fr)
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "new gzip reader failed"))
	}
	defer gzr.Close()

	count := LangCount{}
	reader := gzr
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
				count.ZH++
			case "en":
				count.EN++
			}
		}
	}
	return count
}

func main() {
	cfg, err := lib.LoadDefaultConfig()
	if err != nil {
		lib.Logger.Fatal(errors.Wrap(err, "load config failed"))
	}

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
			count := parseLog(resp.Body)
			err = insertLog(db, date, count)
			if err != nil {
				lib.Logger.Println(err)
			}
		}
	}
}
