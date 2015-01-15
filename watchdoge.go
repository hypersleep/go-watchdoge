package main

import(
	"fmt"
	"net/http"
	"errors"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"os/user"
	"io/ioutil"
	"bytes"
	"strings"
	"strconv"
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
	W http.ResponseWriter
	R *http.Request
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

func ConnectAndRun(params Watchdoge, cmd string) (response string, err error) {
	client, err := ssh.Dial("tcp", params.Server+":22", params.Config)
	if err != nil {
		fmt.Fprintf(params.W, "Failed to dial: %s", err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
		return
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b	
	if err := session.Run(cmd); err != nil {
		panic("Failed to run: " + err.Error())
	}
	response = b.String()
	return 
}

func RunPidstat(params Watchdoge, pid string) {
	cmd := "pidstat -r -p " + pid +" | grep " + params.Procname
	response, err := ConnectAndRun(params, cmd)
	if err != nil {
		panic("Failed to run pidstat: " + err.Error())
	}
	fmt.Println(strings.TrimSpace(response))
	return
}

func FindRemoteProcess(params Watchdoge) (pid string) {
	cmd := "ps ax -o pid,command | grep " + params.Procname + " | grep " + params.Subprocname + " | grep -v bash | awk '{print $1}'"
	response, err := ConnectAndRun(params, cmd)
	if err != nil {
		panic("Failed to find remote process: " + err.Error())
	}
	pid = strings.TrimSpace(response)
	return
}

func PullRemoteProcessMetrics(params Watchdoge) {
	fmt.Println("Process info:")
	pid := FindRemoteProcess(params)
	for i := 0; i < params.Iterations; i++ {
		go RunPidstat(params, pid)
		time.Sleep(time.Duration(params.Period) * time.Second)
	}
}

func CheckSSHConnection(params Watchdoge) (err error) {
	response, err := ConnectAndRun(params, "uname")
	response = strings.TrimSpace(response)
	if err != nil {
		fmt.Fprintf(params.W, "Failed to estabilish SSH connection: %s", err)
	}
	return
}

func ParseParam(r *http.Request, param string) (parsed_param string, err error) {
	parsed_param = r.URL.Query().Get(param)
	if len(parsed_param) == 0 {
		err = errors.New("Ensure " + param)
		return
	}
	return
}

func ParseIntParam(r *http.Request, param string) (parsed_param int, err error) {
	param = r.URL.Query().Get(param)
	parsed_param, _ = strconv.Atoi(param)
	return
}

func GetParams(w http.ResponseWriter, r *http.Request) (params Watchdoge, err error) {
	user, err := ParseParam(r, "user")
	if err != nil {
		fmt.Fprintf(w, "Failed to fetch user parameter!")
		return
	}
	server, err := ParseParam(r, "server")
	if err != nil {
		fmt.Fprintf(w, "Failed to fetch server parameter!")
		return
	}
	procname, err := ParseParam(r, "procname")
	if err != nil {
		fmt.Fprintf(w, "Failed to fetch procname parameter!")
		return
	}
	subprocname, err := ParseParam(r, "subprocname")
	if err != nil {
		fmt.Fprintf(w, "Failed to fetch subprocname parameter!")
		return
	}
	period, err := ParseIntParam(r, "period")
	if err != nil {
		period = 1
	}
	iterations, err := ParseIntParam(r, "iterations")
	if err != nil {
		iterations = 3
	}
	pubkey := getKeyFile()
	config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}
	params = Watchdoge{user, server, procname, subprocname, period, iterations, config, w, r}
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	
	params, err := GetParams(w, r)
	if err != nil { return }

	if err := CheckSSHConnection(params); err != nil { return }

	PullRemoteProcessMetrics(params)

	b, err := json.Marshal(map[string]bool{"success":true})
	if err != nil {
		fmt.Fprintf(w, "Failed to parse json: %s", err)
		return
	}
	fmt.Fprintf(w, string(b))
}

func status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func main() {
	http.HandleFunc("/", status)
	http.HandleFunc("/api", handler)
	http.ListenAndServe(":8080", nil)	
}