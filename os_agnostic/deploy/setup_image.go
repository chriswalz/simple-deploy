package deploy

import "strings"

func (b *Buddy) installDockerOnUbuntu() {
	b.RunCmdsRemotely(
		"sudo apt install docker.io",
		"sudo systemctl start docker",
		"sudo systemctl enable docker",

	)
}

func (b *Buddy) updateUbuntu() {
	b.RunCmdsRemotely(
		"sudo apt-get update && apt-get -y upgrade",
	)
}

func (b *Buddy) setupSupervisorToRemote(path string, appNames ...string) {
	createDirectories := make([]string, 10)
	for _, name := range appNames {
		createDirectories = append(createDirectories, "sudo mkdir -p /var/log/"+name)
	}
	b.RunCmdsRemotely(
		"sudo apt-get install -y supervisor",
		"sudo service supervisor start",
		strings.Join(createDirectories, ";"),
	)
	b.CopySupervisorConfigToRemote(path)
}

func (b *Buddy) CopySupervisorConfigToRemote(srcPath string) {
	b.CopyToRemote(srcPath, "/etc/supervisor/conf.d/sdapps.conf")
	b.RunCmdsRemotely(
		"sudo mkdir -p /var/log/sdapp",
		"supervisorctl reread",
		"supervisorctl update",
	)
}
