package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pawski/proxkeep/application/command"
	"github.com/pawski/proxkeep/application/configuration"
	"github.com/pawski/proxkeep/application/service"
	"github.com/pawski/proxkeep/domain/proxy"
	"github.com/pawski/proxkeep/infrastructure/logger/logrus"
	"github.com/pawski/proxkeep/infrastructure/network/http_client"
	"github.com/pawski/proxkeep/infrastructure/repository"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "ProxKeep"
	app.Version = "v0.1"
	app.Description = "I wish I know it..."
	app.Usage = "Only Go knows"

	app.Before = func(c *cli.Context) error {
		getLogger().Info(app.Name, " - ", app.Version)

		return nil
	}

	appConfig := getAppConfiguration()

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "run",
			Usage: "run [ip] [port]",
			Action: func(c *cli.Context) error {

				envConfig := getEnvConfiguration()
				db, err := sql.Open("mysql", envConfig.MysqlDSN)

				if err != nil {
					getLogger().Error(err)
					return err
				}

				return command.NewRunCommand(
					proxy.NewTester(http_client.NewHttpClient(appConfig.HttpTimeout, getLogger())),
					repository.NewProxyServerRepository(db, getLogger()),
					getLogger(),
					service.NewMeasurement(getLogger())).
					Execute(appConfig.TestUrl, appConfig.ProxyMaxConcurrentChecks)

			},
		}, cli.Command{
			Name:  "test",
			Usage: "test [ip] [port]",
			Action: func(c *cli.Context) error {
				return command.
					NewTestCommand(http_client.NewHttpClient(appConfig.HttpTimeout, getLogger()), getLogger()).
					Execute(appConfig.TestUrl, c.Args().Get(0), c.Args().Get(1))
			},
		}, cli.Command{
			Name:  "selftest",
			Usage: "Takes attempt to fetch test page content",
			Action: func(c *cli.Context) error {
				response, err := http_client.NewHttpClient(appConfig.HttpTimeout, getLogger()).DirectFetch(appConfig.SelfTestUrl)

				fmt.Println(response.StatusCode)
				fmt.Println(string(response.Body))

				return err
			},
		},
	}

	appErr := app.Run(os.Args)

	if appErr != nil {
		getLogger().Fatal(appErr)
	}
}

func getEnvConfiguration() configuration.EnvConfig {
	cfg, err := configuration.GetEnv()

	if err != nil {
		getLogger().Error("Cannot load Env Configuration")
		panic(err)
	}

	return cfg
}

func getAppConfiguration() configuration.AppConfig {
	cfg, err := configuration.GetApp()

	if err != nil {
		getLogger().Errorf("Cannot load App Configuration, using defaults. Cause: %v", err)
		cfg = configuration.GetAppDefaults()
	}

	return cfg
}

func getLogger() proxy.Logger {
	return logrus.Get()
}
