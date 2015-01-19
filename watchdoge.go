package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"strconv"
	"github.com/hypersleep/easyssh"
)

type Watchdoge struct {
	Pid string 
	Period, Iterations int
}

type ServerStream struct {
	Write http.ResponseWriter
	Read *http.Request
}

type Status struct {
	Success bool
}

type Process struct {
	Pid int
	Command string
}

func RunPidstat(ssh_params *easyssh.MakeConfig, watchdoge *Watchdoge) {
	cmd := "LANG=ru_RU.UTF-8 pidstat -r -p " + watchdoge.Pid
	response, _ := ssh_params.ConnectAndRun(cmd)
	fmt.Println(strings.TrimSpace(response))
	return
}

func PullRemoteProcessMetrics(ssh_params *easyssh.MakeConfig, watchdoge *Watchdoge) {
	fmt.Println("Process info:")
	for i := 0; i < watchdoge.Iterations; i++ {
		go RunPidstat(ssh_params, watchdoge)
		time.Sleep(time.Duration(watchdoge.Period) * time.Second)
	}
}

func renderJSON(stream ServerStream, structure interface{}) {
	b, err := json.Marshal(structure)
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to parse json: %s", err)
		return
	}
	fmt.Fprintf(stream.Write, string(b))	
}

func api_handler(w http.ResponseWriter, r *http.Request) {
	stream := ServerStream { w, r }

	pid := r.URL.Query().Get("pid")
	period, _ := strconv.Atoi(r.URL.Query().Get("period"))
	iterations, _ := strconv.Atoi(r.URL.Query().Get("iterations"))

	watchdoge_params := &Watchdoge {
		Pid: pid,
		Period: period,
		Iterations: iterations,
	}

	ssh_params := &easyssh.MakeConfig {
        User: "core",
        Server: "core",
        Key: "/.ssh/id_rsa",
    }

	PullRemoteProcessMetrics(ssh_params, watchdoge_params)

	renderJSON(stream, Status{true})
}

func status(w http.ResponseWriter, r *http.Request) {
	stream := ServerStream { w, r }
	renderJSON(stream, Status{true})
}

func ps(w http.ResponseWriter, r *http.Request) {
	stream := ServerStream { w, r }

	ssh_params := &easyssh.MakeConfig {
        User: "core",
        Server: "core",
        Key: "/.ssh/id_rsa",
    }

	response, err := ssh_params.ConnectAndRun("ps axho pid,command | awk '{print $1\" \"$2}'")
	if err != nil {
		renderJSON(stream, Status{false})
		return
	} else {
		// var processes []Process
		processes := strings.Split(response, "\n")
		renderJSON(stream, processes)
	}
}

func main() {
	port := "8080"
	http.HandleFunc("/status", status)
	http.HandleFunc("/api", api_handler)
	http.HandleFunc("/ps", ps)
	fmt.Println("Server running on port", port)
	http.ListenAndServe(":" + port, nil)
}