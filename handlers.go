package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
	"github.com/hypersleep/easyssh"
	"regexp"
	"time"
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

type Metric struct {
	Key string    `json:"date"`
	Value string  `json:"close"`
}

func metrics(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Requested-With,")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	server_name := r.URL.Query().Get("server")
	server, err := redis_client.Get("servers:" + server_name + ":ip").Result()
	if err != nil {
		renderJSON(w, Status{false})
		return
	}
	// _, metric_keys, err := redis_client.Scan(0, "metrics:" + server + ":*", 50).Result()
	metric_keys, err := redis_client.Keys("metrics:" + server + ":*").Result()
	if err != nil {
		renderJSON(w, Status{false})
		return
	}
	var metrics []Metric
	for _, metric_key := range metric_keys {
		metric_value, err := redis_client.Get(metric_key).Result()
		if err != nil {
			renderJSON(w, Status{false})
			return
		}
		arr := strings.Split(metric_key, ":")
		metric_key = arr[4]
		i, err := strconv.ParseInt(metric_key, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		metric_key = strconv.Itoa(tm.Day())
		metric_key += "-" + tm.Month().String()
		metric_key += "-" + strconv.Itoa(tm.Year())
		metric_key += "-" + strconv.Itoa(tm.Hour())
		metric_key += "-" + strconv.Itoa(tm.Minute())
		metric_key += "-" + strconv.Itoa(tm.Second())
		metric_value = strings.TrimSpace(metric_value)
		metric_value = strings.Replace(metric_value, "kB", "", -1)
		metrics = append(metrics, Metric{metric_key, metric_value})
	}
	renderJSON(w, metrics)
}