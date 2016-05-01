package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/docker/pkg/archive"
	"github.com/fsouza/go-dockerclient"

	"github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

const SEPARATOR = string(filepath.Separator)
const ROOT_FS = "." + SEPARATOR + "rootfs_overlay"

func unpackImage(c *cli.Context) {

	var sourceImage string
	var output string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		output = c.Args()[1]
	} else {
		jww.FATAL.Fatalln("This command requires to argument: source-image output-folder(absolute)")
		os.Exit(1)
	}
	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	if c.GlobalBool("pull") == true {
		PullImage(client, sourceImage)
	}

	if c.Bool("squash") == true {
		jww.INFO.Println("Squashing and unpacking " + sourceImage + " in " + output)
		time := strconv.Itoa(int(makeTimestamp()))
		Squash(client, sourceImage, sourceImage+"-tmpsquashed"+time)
		sourceImage = sourceImage + "-tmpsquashed" + time
		defer func() {
			jww.INFO.Println("Removing squashed image " + sourceImage)
			client.RemoveImage(sourceImage)
		}()
	}

	jww.INFO.Println("Unpacking " + sourceImage + " in " + output)
	Unpack(client, sourceImage, output)
}
func Unpack(client *docker.Client, image string, dirname string) (bool, error) {
	var err error
	r, w := io.Pipe()

	if dirname == "" {
		dirname = ROOT_FS
	}

	os.MkdirAll(dirname, 0777)

	filename, err := ioutil.TempFile(os.TempDir(), "artemide")
	if err != nil {
		jww.FATAL.Fatal("Couldn't create the temporary file")
	}
	os.Remove(filename.Name())

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
		err := client.ExportContainer(docker.ExportContainerOptions{ID: container.ID, OutputStream: w})
		if err != nil {
			jww.FATAL.Fatalln("Couldn't export container, sorry", err)
		}

	}()

	jww.INFO.Println("Extracting to", dirname)

	err = Untar(r, dirname, true)
	if err != nil {
		jww.ERROR.Println("could not unpack to", dirname, err)
		return false, err
	}
	prepareRootfs(dirname)

	return true, err
}

func prepareRootfs(dirname string) {

	err := os.Remove(dirname + SEPARATOR + ".dockerenv")
	if err != nil {
		jww.ERROR.Println("could not remove docker env file")
	}

	err = os.Remove(dirname + SEPARATOR + ".dockerinit")
	if err != nil {
		jww.ERROR.Println("could not remove docker init file")
	}

	err = os.MkdirAll(dirname+SEPARATOR+"dev", 0751)
	if err != nil {
		jww.ERROR.Println("could not create dev folder")
	}

	// Google DNS as default
	d1 := []byte("nameserver 8.8.8.8\nnameserver 8.8.4.4\n")
	err = ioutil.WriteFile(dirname+SEPARATOR+"etc"+SEPARATOR+"resolv.conf", d1, 0644)

}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func Untar(in io.Reader, dest string, sameOwner bool) error {
	return archive.Untar(in, dest, &archive.TarOptions{
		NoLchown:        !sameOwner,
		ExcludePatterns: []string{"dev/"}, // prevent 'operation not permitted'
	})
}
