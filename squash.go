package main

import (
	"github.com/urfave/cli"
	"github.com/mudler/docker-companion/api"
	jww "github.com/spf13/jwalterweatherman"
)

func squashImage(c *cli.Context) error {

	var sourceImage string
	var outputImage string

	client, err := api.NewDocker()
	if err != nil {
		return cli.NewExitError("could not connect to the Docker daemon", 87)
	}

	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		outputImage = c.Args()[1]
		jww.DEBUG.Println("sourceImage " + sourceImage + " outputImage: " + outputImage)
	} else if c.NArg() == 1 {
		sourceImage = c.Args()[0]
		outputImage = sourceImage
		jww.WARN.Println("You didn't specified a second image, i'll squash the one you supplied.")
		if c.Bool("remove") == false {
			jww.WARN.Println("!!! Be careful, docker will leave an image tagged as <none> which is your old one. Use the --remove option to remove it automatically")
		}
		jww.DEBUG.Println("sourceImage " + sourceImage + " outputImage: " + outputImage)
		oldImage, err := client.InspectImage(sourceImage)
		if c.Bool("remove") == true && err == nil {
			defer func(id string) {
				jww.INFO.Println("Removing the untagged image left by the overwrite ID: " + id)
				client.RemoveImage(id)
			}(oldImage.ID)
		}
	} else {
		return cli.NewExitError("This command requires two arguments: squash source-image output-image", 86)
	}

	if c.GlobalBool("pull") == true {
		api.PullImage(client, sourceImage)
	}
	jww.INFO.Println("Squashing " + sourceImage + " in " + outputImage)

	err = api.Squash(client, sourceImage, outputImage)
	if err == nil {
		jww.INFO.Println("Done")
	}
	return err
}
