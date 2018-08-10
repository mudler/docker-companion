package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/archive"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/codegangsta/cli"
	jww "github.com/spf13/jwalterweatherman"
)

// SEPARATOR contains system-specific separator
const SEPARATOR = string(filepath.Separator)

// ROOTFS is our temporary rootfs path
const ROOTFS = "." + SEPARATOR + "rootfs_overlay"

func unpackImage(c *cli.Context) error {

	var sourceImage string
	var output string
	if c.NArg() == 2 {
		sourceImage = c.Args()[0]
		output = c.Args()[1]
	} else {
		return cli.NewExitError("This command requires to argument: source-image output-folder(absolute)", 86)
	}
	client, err := NewDocker()
	if err != nil {
		return cli.NewExitError("could not connect to the Docker daemon", 87)
	}
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
	err = Unpack(client, sourceImage, output, c.GlobalBool("fatal"))
	if err == nil {
		jww.INFO.Println("Done")
	}
	return err
}

// Unpack unpacks a docker image into a path
func Unpack(client *docker.Client, image string, dirname string, fatal bool) error {
	var err error
	r, w := io.Pipe()

	if dirname == "" {
		dirname = ROOTFS
	}

	os.MkdirAll(dirname, 0777)

	filename, err := ioutil.TempFile(os.TempDir(), "artemide")
	if err != nil {
		return cli.NewExitError("Couldn't create the temporary file", 86)
	}
	os.Remove(filename.Name())

	jww.INFO.Println("Creating container")

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
			Cmd:   []string{"true"},
		},
	})
	if err != nil {
		jww.FATAL.Fatalln("Couldn't export container, sorry", err)
	}
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
				jww.WARN.Println("SIGTERM/SIGINT/SIGQUIT detected, removing pending containers")
				client.RemoveContainer(docker.RemoveContainerOptions{
					ID:    container.ID,
					Force: true,
				})
			}
		}
	}()

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
		return cli.NewExitError("could not unpack to "+dirname, 86)
	}
	err = prepareRootfs(dirname, fatal)

	return err
}

func prepareRootfs(dirname string, fatal bool) error {

	_, err := os.Stat(dirname + SEPARATOR + ".dockerenv")
	if err == nil {
		err = os.Remove(dirname + SEPARATOR + ".dockerenv")
		if err != nil {
			if fatal == true {
				return cli.NewExitError("could not remove docker env file", 86)
			} else {
				jww.WARN.Println("error on remove .dockerenv, extracting anyway")
			}
		}
	}

	_, err = os.Stat(dirname + SEPARATOR + ".dockerinit")
	if err == nil {
		err = os.Remove(dirname + SEPARATOR + ".dockerinit")
		if err != nil {
			if fatal == true {
				return cli.NewExitError("could not remove docker init file", 86)
			} else {
				jww.WARN.Println("error on remove .dockerinit, extracting anyway")
			}
		}
	}

	err = os.MkdirAll(dirname+SEPARATOR+"dev", 0751)
	if err != nil {
		if fatal == true {
			return cli.NewExitError("could not create dev folder", 86)
		} else {
			jww.WARN.Println("could not create dev folder")
		}
	}

	// Google DNS as default
	d1 := []byte("nameserver 8.8.8.8\nnameserver 8.8.4.4\n")
	err = ioutil.WriteFile(dirname+SEPARATOR+"etc"+SEPARATOR+"resolv.conf", d1, 0644)
	if err != nil {
		if fatal == true {
			return cli.NewExitError("could not write resolv.conf file", 86)
		} else {
			jww.WARN.Println("could not create resolv.conf file")
		}
	}

	return nil
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// Untar just a wrapper around the docker functions
func Untar(in io.Reader, dest string, sameOwner bool) error {
	return archive.Untar(in, dest, &archive.TarOptions{
		NoLchown:        !sameOwner,
		ExcludePatterns: []string{"dev/"}, // prevent 'operation not permitted'
	})
}
