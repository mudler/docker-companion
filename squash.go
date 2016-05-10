package main

import (
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	jww "github.com/spf13/jwalterweatherman"
)

func squashImage(c *cli.Context) error {

	var sourceImage string
	var outputImage string
	var client *docker.Client
	if os.Getenv("DOCKER_SOCKET") != "" {
		client, _ = docker.NewClient(os.Getenv("DOCKER_SOCKET"))
	} else {
		client, _ = docker.NewClient("unix:///var/run/docker.sock")
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
		PullImage(client, sourceImage)
	}
	jww.INFO.Println("Squashing " + sourceImage + " in " + outputImage)

	err := Squash(client, sourceImage, outputImage)
	return err
}

// Squash Squashes a docker image into another one
func Squash(client *docker.Client, image string, toImage string) error {
	var err error
	var Tag = "latest"
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

	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for {
			s := <-signalchan
			switch s {

			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				jww.WARN.Println("SIGTERM/SIGINT/SIGQUIT detected, removing pending containers/image")
				client.RemoveContainer(docker.RemoveContainerOptions{
					ID:    container.ID,
					Force: true,
				})
				client.RemoveImage(toImage)

			}
		}
	}()

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
		return cli.NewExitError("Could not import docker image", 86)
	}

	return nil
}
