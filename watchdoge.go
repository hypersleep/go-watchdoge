package main

import(
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os/user"
	"io/ioutil"
	"bytes"
	"strings"
	"time"
)


type Watchdoge struct {
	User string
	Server string
	Procname string
	Subprocname string
	Period int
	Iterations int
	Config *ssh.ClientConfig
}

func getKeyFile() (pubkey ssh.Signer){

	usr, err := user.Current()
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	pubkey, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}
	return
}

func ConnectAndRun(params Watchdoge, cmd string) string {
	client, err := ssh.Dial("tcp", params.Server+":22", params.Config)
	if err != nil {
		panic("Failed to dial: "+ err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b	
	if err := session.Run(cmd); err != nil {
		panic("Failed to run: " + err.Error())
	}
	return b.String()
}

func RunPidstat(params Watchdoge, pid string) {
	cmd := "pidstat -r -p " + pid +" | grep " + params.Procname
	fmt.Println(strings.TrimSpace(ConnectAndRun(params, cmd)))
}

func FindRemoteProcess(params Watchdoge) (pid string) {
	cmd := "ps ax -o pid,command | grep " + params.Procname + " | grep " + params.Subprocname + " | grep -v bash | awk '{print $1}'"
	pid = strings.TrimSpace(ConnectAndRun(params, cmd))
	return
}

func PullRemoteProcessMetrics(params Watchdoge) {
	pid := FindRemoteProcess(params)
	for i := 0; i < params.Iterations; i++ {
		go RunPidstat(params, pid)
		if i+1 < params.Iterations {time.Sleep(time.Duration(params.Period) * time.Second)}
	}
}

func main() {

	user := flag.String("user", "root", "SSH username")
	server := flag.String("server", "localhost", "Server name for watchdogging")
	procname := flag.String("procname","ssh","Process name")
	subprocname := flag.String("subprocname", "", "Subprocess name")
	period := flag.Int("period", 1, "Period between statistics checks")
	iterations := flag.Int("iterations", 3, "Number statistics checks")
	
	flag.Parse()

	pubkey := getKeyFile()
	config := &ssh.ClientConfig{
			User: *user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}

	params := Watchdoge{*user, *server, *procname, *subprocname, *period, *iterations, config}
	fmt.Println("Process info:")
    PullRemoteProcessMetrics(params)
    time.Sleep(500 * time.Millisecond)
}
