package main

import(
	"fmt"
	"net/http"
	"errors"
	"encoding/json"
	"golang.org/x/crypto/ssh"	
	"bytes"
	"strings"
	"strconv"
	"time"
	"./test"
)

func ConnectAndRun (params test.Watchdoge, cmd string) (response string, err error) {
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

func RunPidstat (params test.Watchdoge, pid string) {
	cmd := "pidstat -r -p " + pid +" | grep " + params.Procname
	response, err := ConnectAndRun(params, cmd)
	if err != nil {
		panic("Failed to run pidstat: " + err.Error())
	}
	fmt.Println(strings.TrimSpace(response))
	return
}

func FindRemoteProcess (params test.Watchdoge) (pid string) {
	cmd := "ps ax -o pid,command | grep " + params.Procname + " | grep " + params.Subprocname + " | grep -v bash | awk '{print $1}'"
	response, err := ConnectAndRun(params, cmd)
	if err != nil {
		panic("Failed to find remote process: " + err.Error())
	}
	pid = strings.TrimSpace(response)
	return
}

func PullRemoteProcessMetrics (params test.Watchdoge) {
	fmt.Println("Process info:")
	pid := FindRemoteProcess(params)
	for i := 0; i < params.Iterations; i++ {
		go RunPidstat(params, pid)
		time.Sleep(time.Duration(params.Period) * time.Second)
	}
}

func CheckSSHConnection(params test.Watchdoge) (err error) {
	response, err := ConnectAndRun(params, "uname")
	response = strings.TrimSpace(response)
	if err != nil {
		fmt.Fprintf(params.W, "Failed to estabilish SSH connection: %s", err)
	}
	return
}

func renderJSON() {
	b, err := json.Marshal(success)
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to parse json: %s", err)
		return
	}
	fmt.Fprintf(stream.Write, string(b))	
}

func api_handler (w http.ResponseWriter, r *http.Request) {

	const success = map[string]bool{"success":true}
	const fail = map[string]bool{"success":false}

	stream = test.ServerStream {w, r}
	
	params, err := test.Watchdoge.GetParams(stream)
	if err != nil { return }

	ssh_params, err := test.EasySSH.GetConfig(stream)
	if err != nil { return }

	ssh_conn, err := CheckSSHConnection(params)
	if  err != nil { return }

	metrics, err := PullRemoteProcessMetrics(params)
	if  err != nil { return }
}

func status (w http.ResponseWriter, r *http.Request) {
	stream = test.ServerStream {w, r}
	fmt.Fprintf(stream.Write, "ok")
}

func main() {
	http.HandleFunc("/", status)
	http.HandleFunc("/api", api_handler)
	http.ListenAndServe(":8080", nil)	
}