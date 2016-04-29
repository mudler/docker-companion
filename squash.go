package main

import (
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	jww "github.com/spf13/jwalterweatherman"
)

func squashImage(c *cli.Context) {

	var sourceImage string
	var outputImage string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		outputImage = c.Args()[1]
		jww.DEBUG.Println("sourceImage " + sourceImage + " outputImage: " + outputImage)
	} else {
		jww.FATAL.Fatalln("This command requires two arguments: squash source-image output-image")
		os.Exit(1)
	}

	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	if c.GlobalBool("pull") == true {
		PullImage(client, sourceImage)
	}
	jww.INFO.Println("Squashing " + sourceImage + " in " + outputImage)

	Squash(client, sourceImage, outputImage)
}

func Squash(client *docker.Client, image string, toImage string) (bool, error) {
	var err error
	var Tag string = "latest"
	r, w := io.Pipe()

	Imageparts := strings.Split(toImage, ":")
	if len(Imageparts) == 2 {
		Tag = Imageparts[1]
		toImage = Imageparts[0]
	}

	jww.INFO.Println("Creating container")

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
			Cmd:   []string{"true"},
		},
	})
	defer func(*docker.Container) {
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}(container)

	// writing without a reader will deadlock so write in a goroutine
	go func() {
		// it is important to close the writer or reading from the other end of the
		// pipe will never finish
		defer w.Close()
		err = client.ExportContainer(docker.ExportContainerOptions{ID: container.ID, OutputStream: w})
		if err != nil {
			jww.FATAL.Fatalln("Couldn't export container, sorry", err)
		}

	}()

	jww.INFO.Println("Importing to", toImage)

	err = client.ImportImage(docker.ImportImageOptions{Repository: toImage,
		Source:      "-",
		InputStream: r,
		Tag:         Tag,
	})
	if err != nil {
		jww.FATAL.Fatalln("Couldn't import image, sorry", err)
		return false, err
	}

	return true, err
}
