package main

import (
	"github.com/fsouza/go-dockerclient"
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
