package cmd

import (
	"strings"
)

type Config struct {
	port              int
	project           string
	sqlConnectionName string
	sqlInstanceName   string
	dbList            []string
	dbName            string
	dbType            string
	userName          string
}

func New() *Config {
	c := new(Config)
	c.port = 9999
	return c
}

func (c *Config) SetPort(port int) {
	c.port = port
}

func (c *Config) SetProject(project string) {
	c.project = project
}

func (c *Config) SetSqlConnectionName(sqlConnectionName string) {
	c.sqlInstanceName = strings.Split(sqlConnectionName, ":")[2]
	c.sqlConnectionName = sqlConnectionName
}

func (c *Config) SetDbName(dbName string) {
	c.dbName = dbName
}

func (c *Config) SetDbType(dbType string) {
	c.dbType = dbType
}

func (c *Config) SetUserName(userName string) {
	c.userName = userName
}
