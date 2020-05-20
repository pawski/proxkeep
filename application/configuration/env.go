package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var envConfigFile = "env.yml"
var envConfiguration EnvConfig

type EnvConfig struct {
	MysqlDSN string `yaml:"mysql_dsn"`
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
