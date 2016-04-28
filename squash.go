package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	jww "github.com/spf13/jwalterweatherman"
)

func squashImage(ctx *cli.Context) {

	client, _ := docker.NewClient("unix:///var/run/docker.sock")

	if ctx.String("source-image") == "" {
		jww.FATAL.Fatalln("source image not provided, exiting. (see --help) ")
	}
	if ctx.String("output-image") == "" {
		jww.FATAL.Fatalln("output image not provided, exiting. (see --help) ")
	}
	Squash(client, ctx.String("source-image"), ctx.String("output-image"))
	os.Exit(0)

}

func Squash(client *docker.Client, image string, toimage string) (bool, error) {
	var err error

	filename, err := ioutil.TempFile(os.TempDir(), "artemide")
	if err != nil {
		jww.FATAL.Fatal("Couldn't create the temporary file")
	}
	os.Remove(filename.Name())

	// Pulling the image
	jww.INFO.Printf("Pulling the docker image %s\n", image)
	if err := client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{}); err != nil {
		jww.ERROR.Printf("error pulling %s image: %s\n", image, err)
		return false, err
	} else {
		jww.INFO.Println("Image", image, "pulled correctly")
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

	target := fmt.Sprintf("%s.tar", filename.Name())
	jww.DEBUG.Printf("Writing to target %s\n", target)
	writer, err := os.Create(target)
	if err != nil {
		return false, err
	}

	err = client.ExportContainer(docker.ExportContainerOptions{ID: container.ID, OutputStream: writer})
	if err != nil {
		jww.FATAL.Fatalln("Couldn't export container, sorry", err)
		return false, err
	}

	writer.Sync()

	writer.Close()
	jww.INFO.Println("Importing to", toimage)

	err = client.ImportImage(docker.ImportImageOptions{Repository: toimage,
		Source: target,
	})
	if err != nil {
		jww.FATAL.Fatalln("Couldn't import image, sorry", err)
		return false, err
	}

	return true, err
}
