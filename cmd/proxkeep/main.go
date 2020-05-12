package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
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

	app.Commands = []*cli.Command{
		&cli.Command{
			Name:  "run",
			Usage: "run [ip] [port]",
			Action: func(c *cli.Context) error {

				ip := c.Args().Get(0)
				port := c.Args().Get(1)
				logger.Get().Infof("Using %v:%v", ip, port)

				testUrl := "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
				code, response, err := proxy.DirectFetch(testUrl)

				if code != 200 {
					return errors.New("test URL returned non 200 response code")
				}

				if err != nil {
					return err
				}

				logger.Get().Info("Test data acquired")

				pCode, pResponse, _ := proxy.Fetch(ip, port, testUrl)

				if code == pCode && response == pResponse {
					logger.Get().Info("Proxy - OK")
				} else {
					logger.Get().Info("Proxy - NOK")
				}

				return nil
			},
		}, &cli.Command{
			Name:  "selftest",
			Usage: "Takes attempt to fetch test page content",
			Action: func(c *cli.Context) error {
				code, response, err := proxy.DirectFetch("http://ifconfig.io/all.json")

				fmt.Println(code)
				fmt.Println(response)

				return err
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}
