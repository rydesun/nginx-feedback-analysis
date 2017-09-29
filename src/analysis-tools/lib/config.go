package lib

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

type Config struct {
	LogWatch *ConfigLogWatch
	BbsDb    *ConfigBbsDb
	Db       *ConfigDb
	Dir      *ConfigDir
}

type Db struct {
	Host     string
	Port     uint
	Database string
	User     string
	Password string
}

// UrlBase: http://logs.corp.deepin.io/feedback
// FilePattern: feedback.deepin.org_access.log.([0-9]{4}-[0-9]{2}-[0-9]{2}).gz
type ConfigLogWatch struct {
	UrlBase     string
	FilePattern string
	Db
}

type ConfigBbsDb struct {
	Db
}

type ConfigDb struct {
	Db
}

func (c Db) PortStr() string {
	return strconv.Itoa(int(c.Port))
}

type ConfigDir struct {
	ProjectRoot string
	Data        string
	RawLog      string
	Bin         string
}

func (c *ConfigDir) get(dir, name string) string {
	if filepath.IsAbs(dir) {
		return filepath.Join(dir, name)
	}
	return filepath.Join(c.ProjectRoot, dir, name)
}

func (c *ConfigDir) GetData(name string) string {
	return c.get(c.Data, name)
}

func (c *ConfigDir) GetRawLog(name string) string {
	return c.get(c.RawLog, name)
}

func (c *ConfigDir) GetBin(name string) string {
	return c.get(c.Bin, name)
}

func LoadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadDefaultConfig() (*Config, error) {
	return LoadConfig("./config/app.json")
}
