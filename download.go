package main

import (
	"github.com/codegangsta/cli"
	"github.com/mudler/docker-companion/api"
)

func downloadImage(c *cli.Context) error {

	var sourceImage string
	var output string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		output = c.Args()[1]
	} else {
		return cli.NewExitError("This command requires to argument: source-image output-folder(absolute)", 86)
	}

	return api.DownloadImage(sourceImage, output, "")
}
