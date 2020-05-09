package main

import (
	"crypto/md5"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pawski/proxkeep/logger"
	"github.com/pawski/proxkeep/validator"
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
			Usage: "Go Gadget Go",
			Action: func(c *cli.Context) error {
				code, expectedResponse, err := validator.DirectFetch("https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf")

				if 200 == code {
					validator.HealthCheck("95.79.36.55", "44861", expectedResponse)
				}

				return err
			},
		}, &cli.Command{
			Name:  "selftest",
			Usage: "Takes attempt to fetch test page content",
			Action: func(c *cli.Context) error {
				code, response, err := validator.DirectFetch("http://ifconfig.io/all.json")

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
