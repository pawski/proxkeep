package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var envConfigFile = "env.yml"
var envConfiguration EnvConfig

type EnvConfig struct {
	MysqlDSN         string `yaml:"mysql_dsn"`
	InfluxDbHost     string `yaml:"influx_host"`
	InfluxDbUser     string `yaml:"influx_user"`
	InfluxDbPassword string `yaml:"influx_password"`
	InfluxDbDatabase string `yaml:"influx_database"`
}

func GetEnv() (EnvConfig, error) {
	file, err := ioutil.ReadFile(envConfigFile)

	if err != nil {
		return envConfiguration, err
	}

	err = yaml.Unmarshal(file, &envConfiguration)

	if err != nil {
		return envConfiguration, err
	}

	return envConfiguration, nil
}
