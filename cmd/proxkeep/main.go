package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pawski/proxkeep/application/command"
	"github.com/pawski/proxkeep/infrastructure/logger/logrus"
	"github.com/pawski/proxkeep/infrastructure/network/http"
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
		logrus.Get().Info(app.Name, " - ", app.Version)

		return nil
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "run",
			Usage: "run [ip] [port]",
			Action: func(c *cli.Context) error {

				db, err := sql.Open("mysql", "root:vagrant@tcp(192.168.55.102)/hrs")

				if err != nil {
					logrus.Get().Errorln(err)
					return err
				}

				return command.NewRunCommand(
					repository.NewProxyServerRepository(db, logrus.Get()),
					logrus.Get()).
					Execute()

			},
		}, cli.Command{
			Name:  "test",
			Usage: "test [ip] [port]",
			Action: func(c *cli.Context) error {
				return command.NewTestCommand(logrus.Get()).Execute(c.Args().Get(0), c.Args().Get(1))
			},
		}, cli.Command{
			Name:  "selftest",
			Usage: "Takes attempt to fetch test page content",
			Action: func(c *cli.Context) error {
				response, err := http.DirectFetch("http://ifconfig.io/all.json")

				fmt.Println(response.StatusCode)
				fmt.Println(string(response.Body))

				return err
			},
		},
	}

	appErr := app.Run(os.Args)

	if appErr != nil {
		logrus.Get().Fatal(appErr)
	}
}
