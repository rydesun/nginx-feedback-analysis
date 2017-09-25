package log_analysis

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

type Config struct {
	LogWatch *ConfigLogWatch
	Db       *ConfigDb
	Dir      *ConfigDir
}

// UrlBase: http://logs.corp.deepin.io/feedback
// FilePattern: feedback.deepin.org_access.log.([0-9]{4}-[0-9]{2}-[0-9]{2}).gz
type ConfigLogWatch struct {
	UrlBase     string
	FilePattern string
}

type ConfigDb struct {
	Host     string
	Port     uint
	Database string
	User     string
	Password string
}

func (c *ConfigDb) PortStr() string {
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
