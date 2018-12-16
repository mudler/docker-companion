package api

import (
	"os"
	"os/exec"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

func extractTar(src, dest string) ([]byte, error) {
	jww.INFO.Printf("Extracting: ", TarCmd, "--same-owner", "--xattrs", "--overwrite",
		"--preserve-permissions", "-xf", src, "-C", dest)
	cmd := exec.Command(TarCmd, "--same-owner", "--xattrs", "--overwrite",
		"--preserve-permissions", "-xf", src, "-C", dest)
	return cmd.CombinedOutput()
}

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
	var err error
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
