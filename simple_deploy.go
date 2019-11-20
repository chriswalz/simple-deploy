package main

import (
	"github.com/chriswalz/simple-deploy/os_agnostic/deploy"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"strings"
)

// easiest way to daemonize a go app?
// update readme with gif on how to use
// update readme with setup guide

func main() {
	log.SetFlags(log.Lshortfile)
	var binaryName = "app"
	var appName = "sdapp"

	// todo add "STATUS" section
	app := cli.NewApp()
	app.Name = "Simple Deploy"
	app.Usage = "Fast and simple deploy to unix server"
	app.Version = "0.7.1"
	app.Action = func(c *cli.Context) error {
		if c.Args().Len() < 2 {
			return cli.NewExitError("Error: missing host address\nUsage: simple-deploy <host> <filePaths>\nExample: simple-deploy joe@example.com dir/main.go,static/files", 1)
		}
		if c.Args().Get(0) == "logs" {
			if c.Args().Len() < 2 {
				return cli.NewExitError("Error: missing host address\nUsage: simple-deploy logs <host>\nExample: simple-deploy logs joe@example.com", 1)
			}
			exec.Command("supervisorctl tail -5000 goapp stdout; supervisorctl tail -5000 goapp stderr")
			return nil
		}
		if c.Args().Get(0) == "upload-supervisor-config" {
			if c.Args().Len() < 3 {
				return cli.NewExitError("Error: missing host address\nUsage: simple-deploy upload-supervisor-config <host> <config>\nExample: simple-deploy upload-supervisor-config joe@example.com configs/app.conf", 1)
			}
			user, address := deploy.GetSSHArgs(c.Args().Get(1))
			buddy := deploy.SetupClient(user, address)
			buddy.CopySupervisorConfigToRemote(c.Args().Get(2))
			return nil
		}
		user, address := deploy.GetSSHArgs(c.Args().Get(0))
		buddy := deploy.SetupClient(user, address)
		paths := strings.Split(c.Args().Get(1), ",")

		buddy.Deploy(binaryName, appName, user, address, paths)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
