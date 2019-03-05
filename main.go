package main

import (
	"os"

	"github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

// VERSION is the app version
const VERSION = "0.4.3"

func main() {
	app := cli.NewApp()
	app.Name = "docker-companion"
	app.Usage = "a Candy mix of Docker tools"
	app.Version = VERSION
	jww.SetStdoutThreshold(jww.LevelInfo)
	if os.Getenv("DEBUG") == "1" {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "pull",
			Usage: "pull image before doing operations",
		},
		cli.BoolFlag{
			Name:  "fatal",
			Usage: "threat errors as fatal",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl"},
			Usage:   "Download and unpacks an image without using docker - Usage: download foo/barimage /foobar/folder",
			Action:  downloadImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "keep",
					Usage: "Keeps downloaded layers around (useful for debugging)",
				},
			},
		},
		{
			Name:    "unpack",
			Aliases: []string{"un"},
			Usage:   "unpack the specified Docker image content as-is (run as root!) in a folder - Usage: unpack foo/barimage /foobar/folder",
			Action:  unpackImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "squash",
					Usage: "squash image before doing operations",
				},
			},
		},
		{
			Name:    "squash",
			Aliases: []string{"s"},
			Usage:   "squash the Docker image (loosing metadata) into another - Usage: squash foo/bar foo/bar-squashed:latest. The second argument is optional",
			Action:  squashImage,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "remove, rm",
					Usage: "If you supplied just one image, remove the untagged image",
				},
			},
		},
	}
	jww.INFO.Println("Docker Companion v" + VERSION)

	app.Run(os.Args)
}
