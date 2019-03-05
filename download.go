package main

import (
	"os"

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
	unpackmode := os.Getenv("UNPACK_MODE")
	if unpackmode == "" {
		unpackmode = "umoci"
	}
	return api.DownloadAndUnpackImage(sourceImage, output, &api.DownloadOpts{KeepLayers: c.Bool("keep"), UnpackMode: unpackmode})
}
