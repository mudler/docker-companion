package main

import (
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/mudler/docker-companion/vendor/github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

// PullImage pull the specified image
func PullImage(client *docker.Client, image string) error {
	var err error
	// Pulling the image
	jww.INFO.Printf("Pulling the docker image %s\n", image)
	if err = client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{}); err != nil {
		jww.ERROR.Printf("error pulling %s image: %s\n", image, err)
		return err
	}

	jww.INFO.Println("Image", image, "pulled correctly")

	return nil
}

// NewDocker Creates a new instance of *docker.Client, respecting env settings
func NewDocker() (*docker.Client, error) {
	var client *docker.Client
	if os.Getenv("DOCKER_SOCKET") != "" {
		client, err = docker.NewClient(os.Getenv("DOCKER_SOCKET"))
	} else {
		client, err = docker.NewClient("unix:///var/run/docker.sock")
	}
	if err != nil {
		return nil, cli.NewExitError("could not connect to the Docker daemon", 87)
	}
	return client, nil
}
