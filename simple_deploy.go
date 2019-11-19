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

func main() {
	var binaryName = "app"

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
		user, address := deploy.GetSSHArgs(c.Args().Get(0))
		buddy := deploy.SetupClient(user, address)
		paths := strings.Split(c.Args().Get(1), ",")

		buddy.Deploy(binaryName, user, address, true, paths)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
