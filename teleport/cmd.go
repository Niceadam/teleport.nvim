package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var socket string

func start(ctx *cli.Context) error {
	var name string
	if name = ctx.Args().Get(0); name == "" {
		name = "main"
	}

	socket = "tmp/nvim.sock." + name
	if socket == "" {
		startNvimInstance()
		listen(socket)
	}

	// Run TUI
	cmd := exec.Command("nvim", "--server", socket, "--remote-ui")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	check(err)
	return err
}

func main() {
	// Turn off timestamps in output.
	log.SetFlags(0)

	app := &cli.App{
		Name:   "teleport",
		Usage:  "Share nvim instances",
		Action: start,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
