package dbservice

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type DBsettings struct {
	host     string `yaml:"host"`
	port     int    `yaml:"port"`
	user     string `yaml:"user"`
	password string `yaml:"password"`
	dbname   string `yaml:"dbname"`
}

type internalDBsettings struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"dbname"`
}

func New(path string) *DBsettings {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var s internalDBsettings

	err2 := yaml.Unmarshal(data, &s)
	if err2 != nil {
		log.Fatalf("error: %v", err)
	}

	return &DBsettings{
		host:     s.Host,
		port:     s.Port,
		user:     s.User,
		password: s.Password,
		dbname:   s.DBname,
	}
}

func Host(s *DBsettings) string {
	return s.host
}
func Port(s *DBsettings) int {
	return s.port
}
func User(s *DBsettings) string {
	return s.user
}
func Password(s *DBsettings) string {
	return s.password
}
func DBname(s *DBsettings) string {
	return s.dbname
}
