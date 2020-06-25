package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var appConfigFile = "app.yml"
var appConfiguration AppConfig

type AppConfig struct {
	TestUrl                  string `yaml:"test_url"`
	SelfTestUrl              string `yaml:"self_test_url"`
	ProxyMaxConcurrentChecks uint   `yaml:"proxy_max_concurrent_checks"`
	HttpTimeout              uint   `yaml:"http_timeout"`
	HttpStatsEnabled         bool   `yaml:"enable_http_stats"`
	MetricsCollectorEnabled  bool   `yaml:"enable_metrics_collector"`
}

func GetApp() (AppConfig, error) {
	file, err := ioutil.ReadFile(appConfigFile)

	if err != nil {
		return appConfiguration, err
	}

	err = yaml.Unmarshal(file, &appConfiguration)

	if err != nil {
		return appConfiguration, err
	}

	return appConfiguration, nil
}

func GetAppDefaults() AppConfig {
	return AppConfig{
		SelfTestUrl:              "http://ifconfig.io/all.json",
		TestUrl:                  "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf",
		ProxyMaxConcurrentChecks: 10,
		HttpTimeout:              10,
		HttpStatsEnabled:         false,
		MetricsCollectorEnabled:  false,
	}
}
