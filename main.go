package main

import (
	"os"

	"github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

// VERSION is the app version
const VERSION = "0.1"

func main() {
	app := cli.NewApp()
	app.Name = "docker-companion"
	app.Usage = "a Candy mix of Docker tools"
	app.Version = VERSION
	jww.SetStdoutThreshold(jww.LevelInfo)
	if os.Getenv("DEBUG") == "1" {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}

	app.Commands = []cli.Command{
		{
			Name:    "unpack",
			Aliases: []string{"un"},
			Usage:   "unpack a Docker image content as-is",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "source-image",
					Usage: "Docker source image name",
				},
				cli.StringFlag{
					Name:  "output",
					Usage: "output folder",
				},
			},
			Action: unpackImage,
		},
		{
			Name:    "squash",
			Aliases: []string{"s"},
			Usage:   "squash the Docker image layers (loosing metadata)",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "source-image",
					Usage: "Docker source image name",
				},
				cli.StringFlag{
					Name:  "output-image",
					Usage: "Docker output image name",
				},
			},
			Action: squashImage,
		},
	}

	app.Run(os.Args)
}
