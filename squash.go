package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func squashImage(ctx *cli.Context) {
	fmt.Printf("The first argument was: %s", ctx.Args().First())
}
