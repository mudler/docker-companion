package main

import (
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/mudler/docker-companion/api"
	jww "github.com/spf13/jwalterweatherman"
)

func unpackImage(c *cli.Context) error {

	var sourceImage string
	var output string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		output = c.Args()[1]
	} else {
		return cli.NewExitError("This command requires to argument: source-image output-folder(absolute)", 86)
	}
	client, err := api.NewDocker()
	if err != nil {
		return cli.NewExitError("could not connect to the Docker daemon", 87)
	}
	if c.GlobalBool("pull") == true {
		api.PullImage(client, sourceImage)
	}

	if c.Bool("squash") == true {
		jww.INFO.Println("Squashing and unpacking " + sourceImage + " in " + output)
		time := strconv.Itoa(int(makeTimestamp()))
		api.Squash(client, sourceImage, sourceImage+"-tmpsquashed"+time)
		sourceImage = sourceImage + "-tmpsquashed" + time
		defer func() {
			jww.INFO.Println("Removing squashed image " + sourceImage)
			client.RemoveImage(sourceImage)
		}()
	}

	jww.INFO.Println("Unpacking " + sourceImage + " in " + output)
	err = api.Unpack(client, sourceImage, output, c.GlobalBool("fatal"))
	if err == nil {
		jww.INFO.Println("Done")
	}
	return err
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
