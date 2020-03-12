package deploy

import (
	"compress/flate"
	"fmt"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Buddy struct {
	*ssh.Client
}

func (b *Buddy) Deploy(BinaryName, appName, user, address string, paths []string) {
	var err error

	now := time.Now()
	rand.Seed(time.Now().UTC().UnixNano())
	nowFormatted := now.Format("2006.1.") + strconv.Itoa(rand.Intn(9000)+1000)

	b.BuildAndZip(paths[0], paths[1:], BinaryName, nowFormatted)

	binaryZip := BinaryName + ".zip"
	b.CopyToRemote(binaryZip, "../usr/local/bin/"+binaryZip)
	b.RunCmdsRemotely(
		"cd /usr/local/bin/",
		fmt.Sprintf("unzip -qq -o %s.zip", BinaryName),
		fmt.Sprintf("chmod u+x %s", BinaryName),
		fmt.Sprintf("supervisorctl restart %v", appName),
	)

	resp, err := http.Get("https://" + address)
	if err != nil {
		logs := b.RunCmdRemotelyGetOutput("supervisorctl tail -5000 goapp stderr; supervisorctl status")
		fmt.Println(logs)
		if strings.Contains(logs, "Exited") {
			log.Println(logs)
		}
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("For %s@%s, expected status: %v, got: %v", user, address, http.StatusOK, resp.StatusCode)
	}

	Run("git", "tag", nowFormatted)
	Run("git", "push", "origin", nowFormatted)

}


// NewBuddy
func SetupClient(user, address string) *Buddy {
	home, err := os.UserHomeDir()
	encryptedKey, err := ioutil.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		log.Fatal(err)
	}

	// ssh private key password
	fmt.Print("Enter ssh key passphrase: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Println(err)
	}
	pw := string(bytePassword)
	fmt.Println()
	signer, err := ssh.ParsePrivateKeyWithPassphrase(encryptedKey, []byte(pw))
	if err != nil {
		log.Fatal(err)
	}

	client, err := ssh.Dial("tcp", address + ":22",  &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Timeout: time.Duration(20) * time.Second,
	})

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Connected to server.")
		return &Buddy{
			client,
		}
	}

	// if ssh key fails try using password
	fmt.Print("Enter server password: ")
	bytePassword, err = terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Println(err)
	}
	pw = string(bytePassword)
	fmt.Println()

	client, err = ssh.Dial("tcp", address+":22", &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pw),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(20) * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to dial: %s@%s, error: %s", user, address, err.Error())
	}
	return &Buddy{
		client,
	}
}

func (b *Buddy) RunCmdRemotelyGetOutput(cmdStr string) string {
	var err error

	conn, err := b.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	out, err := conn.CombinedOutput(cmdStr)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		if !ok {
			log.Fatal("couldn't get caller")
		}
		fmt.Printf("%s#%d\n", file, no)
		log.Fatal(string(out), err)
	}
	return string(out)
}

func (b *Buddy) RunCmdRemotely(cmdStr string) {
	var err error
	conn, err := b.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	out, err := conn.CombinedOutput(cmdStr)
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		if !ok {
			log.Fatal("couldn't get caller")
		}
		fmt.Printf("%s#%d\n", file, no)
		log.Fatal(string(out), err)
	}
}

func (b *Buddy) RunCmdsRemotely(cmds ...string) {
	var err error
	conn, err := b.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	out, err := conn.CombinedOutput(strings.Join(cmds, ";"))
	if err != nil {
		_, file, no, ok := runtime.Caller(1)
		if !ok {
			log.Fatal("couldn't get caller")
		}
		fmt.Printf("%s#%d\n", file, no)
		log.Fatal(string(out), err)
	}
}

func (b *Buddy) CopyToRemote(srcPath, dstPath string) {
	if !strings.Contains(dstPath, ".") {
		log.Fatal("destination path must also include the name of the file that will be created", dstPath)
	}

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(b.Client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Fatal(err, ": ", srcPath)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// Copy the file
	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Buddy) BuildAndZip(mainFilePath string, additionalPaths []string, binaryName, version string) {
	var err error
	log.SetFlags(log.Lshortfile)

	zipName := binaryName + ".zip"
	os.Remove(zipName)

	cmd := exec.Command("go", "build", "-o", binaryName, mainFilePath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux")

	err = cmd.Run()
	if err != nil {
		log.Println("failed run")
		log.Fatal(err)
	}

	z := archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      false,
		ImplicitTopLevelFolder: false,
	}

	archivePaths := append(additionalPaths, binaryName)

	err = z.Archive(archivePaths, zipName)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(binaryName)
	if err != nil {
		log.Fatal(err)
	}
}

func GetSSHArgs(url string) (string, string) {
	if !strings.Contains(url, "@") {
		log.Fatal("must follow user@server.com")
	}
	user := strings.Split(url, "@")[0]
	address := strings.Split(url, "@")[1]

	return user, address
}
