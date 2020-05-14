package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pawski/proxkeep/application/command"
	"github.com/pawski/proxkeep/infrastructure/repository"
	"github.com/pawski/proxkeep/logger"
	"github.com/pawski/proxkeep/proxy"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "ProxKeep"
	app.Version = "v0.1"
	app.Description = "I wish I know it..."
	app.Usage = "Only Go knows"

	app.Before = func(c *cli.Context) error {
		logger.Get().Formatter = &logrus.TextFormatter{FullTimestamp: true}
		logger.Get().Info(app.Name, " - ", app.Version)

		return nil
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:  "run",
			Usage: "run [ip] [port]",
			Action: func(c *cli.Context) error {

				db, err := sql.Open("mysql", "root:vagrant@tcp(192.168.55.102)/hrs")

				if err != nil {
					logger.Get().Errorln(err)
					return err
				}

				return command.NewRunCommand(repository.NewProxyServerRepository(db, logger.Get()), logger.Get()).Execute()

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
					logger.Get().Error("IP Address cannot be empty")

					return nil
				}

				logger.Get().Infof("Using %v:%v", ip, port)

				testUrl := "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
				response, err := proxy.DirectFetch(testUrl)

				if response.StatusCode != 200 {
					return errors.New("test URL returned non 200 response code")
				}

				if err != nil {
					return err
				}

				logger.Get().Info("Test data acquired")
				logger.Get().Infof("Main connection Throughput %.4f KB/s", float64(len(response.Body)/1024)/response.TransferTime)

				pResponse, _ := proxy.Fetch(ip, port, testUrl)

				if response.StatusCode == pResponse.StatusCode && 0 == bytes.Compare(response.Body, pResponse.Body) {
					logger.Get().Info("Proxy - OK")
					logger.Get().Infof("Proxy throughput %.4f KB/s", float64(len(pResponse.Body)/1024)/pResponse.TransferTime)
				} else {
					logger.Get().Info("Proxy - NOK")
				}

				return nil
			},
		}, cli.Command{
			Name:  "selftest",
			Usage: "Takes attempt to fetch test page content",
			Action: func(c *cli.Context) error {
				response, err := proxy.DirectFetch("http://ifconfig.io/all.json")

				fmt.Println(response.StatusCode)
				fmt.Println(string(response.Body))

				return err
			},
		},
	}

	appErr := app.Run(os.Args)

	if appErr != nil {
		log.Fatal(appErr)
	}
}
