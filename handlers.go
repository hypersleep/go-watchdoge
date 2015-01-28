package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
	"github.com/hypersleep/easyssh"
	"regexp"
)

type Status struct {
	Success bool
}

type Process struct {
	Pid string
	Command string
}

func ParseProcesses(stdout string) (processes []Process) {
	stdout_parsed := strings.Split(stdout, "\n")
	for _, process := range stdout_parsed {
		r, _ := regexp.Compile("(?:\\s*)([0-9]*)(?:\\s)(.*)")
		process_item := r.FindAllStringSubmatch(process, -1)
		if process_item != nil && process_item[0][1] != "" && process_item[0][2] != "" {
			processes = append(processes, Process{process_item[0][1], process_item[0][2]})
		}
	}
	return
}

func renderJSON(w http.ResponseWriter, structure interface{}) {
	b, err := json.Marshal(structure)
	if err != nil {
		fmt.Fprintf(w, "Failed to parse json: %s", err)
		return
	}
	fmt.Fprintf(w, string(b))	
}

func api_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	server_name := r.URL.Query().Get("server")

	server, _ := redis_client.Get("servers:" + server_name + ":ip").Result()
	user, _ := redis_client.Get("servers:" + server_name + ":ssh_user").Result()

	pid := r.URL.Query().Get("pid")
	period, _ := strconv.Atoi(r.URL.Query().Get("period"))
	iterations, _ := strconv.Atoi(r.URL.Query().Get("iterations"))
	stat := r.URL.Query().Get("stat")

	watchdoge_params := &Watchdoge {
		Pid: pid,
		Period: period,
		Iterations: iterations,
		Stat: stat,
	}

	ssh_params := &easyssh.MakeConfig {
        User: user,
        Server: server,
        Key: "/.ssh/id_rsa",
    }

	go PullRemoteProcessMetrics(ssh_params, watchdoge_params)

	renderJSON(w, Status{true})
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	renderJSON(w, Status{true})
}

func ps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	server_name := r.URL.Query().Get("server")
	server, err := redis_client.Get("servers:" + server_name + ":ip").Result()
	user, err := redis_client.Get("servers:" + server_name + ":ssh_user").Result()

	ssh_params := &easyssh.MakeConfig {
        User: user,
        Server: server,
        Key: "/.ssh/id_rsa",
    }

	response, err := ssh_params.ConnectAndRun("ps axho pid,command --sort rss")
	if err != nil {
		renderJSON(w, Status{false})
		return
	} else {
		processes := ParseProcesses(response)
		renderJSON(w, processes)
	}
}