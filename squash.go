package main

import (
	"os"

	docker "github.com/mudler/docker-companion/docker"
	jww "github.com/spf13/jwalterweatherman"

	"github.com/codegangsta/cli"
)

func squashImage(ctx *cli.Context) {

	client, _ := docker.NewClient("unix:///var/run/docker.sock")

	if ctx.String("source-image") == "" {
		jww.FATAL.Fatalln("source image not provided, exiting. (see --help) ")
	}
	if ctx.String("output-image") == "" {
		jww.FATAL.Fatalln("output image not provided, exiting. (see --help) ")
	}
	client.Squash(ctx.String("source-image"), ctx.String("output-image"))
	os.Exit(0)

}
