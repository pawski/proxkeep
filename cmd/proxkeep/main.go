package main

import (
	"bytes"
	"database/sql"
	"errors"
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

				ip := c.Args().Get(0)
				port := c.Args().Get(1)

				if "" == port {
					port = "8080"
				}

				if "" == ip {
					logrus.Get().Error("IP Address cannot be empty")

					return nil
				}

				logrus.Get().Infof("Using %v:%v", ip, port)

				testUrl := "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
				response, err := http.DirectFetch(testUrl)

				if response.StatusCode != 200 {
					return errors.New("test URL returned non 200 response code")
				}

				if err != nil {
					return err
				}

				logrus.Get().Info("Test data acquired")
				logrus.Get().Infof("Main connection Throughput %.4f KB/s", float64(len(response.Body)/1024)/response.TransferTime)

				pResponse, _ := http.Fetch(ip, port, testUrl)

				if response.StatusCode == pResponse.StatusCode && 0 == bytes.Compare(response.Body, pResponse.Body) {
					logrus.Get().Info("Proxy - OK")
					logrus.Get().Infof("Proxy throughput %.4f KB/s", float64(len(pResponse.Body)/1024)/pResponse.TransferTime)
				} else {
					logrus.Get().Info("Proxy - NOK")
				}

				return nil
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
