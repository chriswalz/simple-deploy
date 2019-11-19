package deploy

import (
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strconv"
	"time"
)

func Run(name string, arg ...string) {
	fmt.Print(name, arg, "..")
	out, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		log.Fatalf("%s %s: %s %s", name, arg, out, err)
	}
	log.Println(string(out))
}

func GenerateVersion() string {
	now := time.Now()
	rand.Seed(time.Now().UTC().UnixNano())
	return now.Format("2006.1.") + strconv.Itoa(rand.Intn(9000)+1000)
}