package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"golang.org/x/crypto/ssh"	
	"bytes"
	"strings"
	"time"
	"./test"
)

func ConnectAndRun (ssh_params test.EasySSH, cmd string) (response string, err error) {
	client, err := ssh.Dial("tcp", ssh_params.Server+":22", ssh_params.Config)
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

func RunPidstat (ssh_params test.EasySSH, watchdoge test.Watchdoge, pid string) {
	cmd := "LANG=ru_RU.UTF-8 pidstat -r -p " + pid +" | grep " + watchdoge.Procname
	response, _ := ConnectAndRun(ssh_params, cmd)
	fmt.Println(strings.TrimSpace(response))
	return
}

func FindRemoteProcess (ssh_params test.EasySSH, watchdoge test.Watchdoge) (pid string) {
	cmd := "ps ax -o pid,command | grep " + watchdoge.Procname + " | grep " + watchdoge.Subprocname + " | grep -v bash | awk '{print $1}'"
	response, err := ConnectAndRun(ssh_params, cmd)
	if err != nil {
		panic("Failed to find remote process: " + err.Error())
	}
	pid = strings.TrimSpace(response)
	return
}

func PullRemoteProcessMetrics (ssh_params test.EasySSH, watchdoge test.Watchdoge) {
	fmt.Println("Process info:")
	pid := FindRemoteProcess(ssh_params, watchdoge)
	for i := 0; i < watchdoge.Iterations; i++ {
		go RunPidstat(ssh_params, watchdoge, pid)
		time.Sleep(time.Duration(watchdoge.Period) * time.Second)
	}
}

func CheckSSHConnection(ssh_params test.EasySSH) {
	response, _ := ConnectAndRun(ssh_params, "uname")
	response = strings.TrimSpace(response)
	fmt.Println(response)
}

func renderJSON (stream test.ServerStream, structure interface{}) {
	b, err := json.Marshal(structure)
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to parse json: %s", err)
		return
	}
	fmt.Fprintf(stream.Write, string(b))	
}

type Status struct {
		Success bool
	}

func api_handler (w http.ResponseWriter, r *http.Request) {

	stream := test.ServerStream {w, r}

	watchdoge := test.Watchdoge {}

	watchdoge.GetProcname(stream)
	watchdoge.GetSubprocname(stream)
	watchdoge.GetPeriod(stream)
	watchdoge.GetProcname(stream)
	watchdoge.GetIterations(stream)

	ssh_params  := test.EasySSH{}

	ssh_params.GetUser(stream)	
	ssh_params.GetConfig(stream)

	CheckSSHConnection(ssh_params)

	PullRemoteProcessMetrics(ssh_params, watchdoge)

	renderJSON(stream, Status{true})
}

func status (w http.ResponseWriter, r *http.Request) {

	stream := test.ServerStream {w, r}

	renderJSON(stream, Status{true})
}

func main() {
	port := "8080"
	http.HandleFunc("/status", status)
	http.HandleFunc("/api", api_handler)
	fmt.Println("Server running on port", port)
	http.ListenAndServe(":" + port, nil)
}