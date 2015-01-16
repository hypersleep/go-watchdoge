package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"./test"
	"./easyssh"
)

type ServerStream struct {
	Write http.ResponseWriter
	Read *http.Request
}

func RunPidstat (ssh_params easyssh.EasySSH, watchdoge test.Watchdoge, pid string) {
	cmd := "LANG=ru_RU.UTF-8 pidstat -r -p " + pid +" | grep " + watchdoge.Procname
	response, _ := ssh_params.ConnectAndRun(cmd)
	fmt.Println(strings.TrimSpace(response))
	return
}

func FindRemoteProcess (ssh_params easyssh.EasySSH, watchdoge test.Watchdoge) (pid string) {
	cmd := "ps ax -o pid,command | grep " + watchdoge.Procname + " | grep " + watchdoge.Subprocname + " | grep -v bash | awk '{print $1}'"
	response, err := ssh_params.ConnectAndRun(cmd)
	if err != nil {
		panic("Failed to find remote process: " + err.Error())
	}
	pid = strings.TrimSpace(response)
	return
}

func PullRemoteProcessMetrics (ssh_params easyssh.EasySSH, watchdoge test.Watchdoge) {
	fmt.Println("Process info:")
	pid := FindRemoteProcess(ssh_params, watchdoge)
	for i := 0; i < watchdoge.Iterations; i++ {
		go RunPidstat(ssh_params, watchdoge, pid)
		time.Sleep(time.Duration(watchdoge.Period) * time.Second)
	}
}

func renderJSON (stream ServerStream, structure interface{}) {
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

	stream := ServerStream {w, r}

	doge_stream := test.ServerStream {w, r}

	watchdoge := test.Watchdoge {}

	watchdoge.GetProcname(doge_stream)
	watchdoge.GetSubprocname(doge_stream)
	watchdoge.GetPeriod(doge_stream)
	watchdoge.GetProcname(doge_stream)
	watchdoge.GetIterations(doge_stream)

	ssh_stream := easyssh.ServerStream {w, r}

	ssh_params  := easyssh.EasySSH {}

	ssh_params.GetUser(ssh_stream)	
	ssh_params.GetConfig(ssh_stream)

	ssh_params.CheckSSHConnection()

	PullRemoteProcessMetrics(ssh_params, watchdoge)

	renderJSON(stream, Status{true})
}

func status (w http.ResponseWriter, r *http.Request) {

	stream := ServerStream {w, r}

	renderJSON(stream, Status{true})
}

func main() {
	port := "8080"
	http.HandleFunc("/status", status)
	http.HandleFunc("/api", api_handler)
	fmt.Println("Server running on port", port)
	http.ListenAndServe(":" + port, nil)
}