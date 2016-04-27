package main

import (
	"os"

	"github.com/codegangsta/cli"
	docker "github.com/mudler/docker-companion/docker"
	jww "github.com/spf13/jwalterweatherman"
)

func unpackImage(ctx *cli.Context) {

	client, _ := docker.NewClient("unix:///var/run/docker.sock")

	if ctx.String("source-image") == "" {
		jww.FATAL.Fatalln("source image not provided, exiting. (see --help) ")
	}
	if ctx.String("output") == "" {
		jww.FATAL.Fatalln("source image not provided, exiting. (see --help) ")
	}
	jww.INFO.Println("Unpacking " + ctx.String("source-image") + " in " + ctx.String("output"))
	client.Unpack(ctx.String("source-image"), ctx.String("output"))
	os.Exit(0)

}
