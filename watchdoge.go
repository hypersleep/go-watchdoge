package main

import(
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"strconv"
	"github.com/hypersleep/easyssh"
	"regexp"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"gopkg.in/redis.v2"
)

type Watchdoge struct {
	Pid, Stat string 
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
	Pid string
	Command string
}

type ServersConfig struct {
	Servers map[string][]string
}

var servers = ServersConfig{}

func RunPidstat(ssh_params *easyssh.MakeConfig, watchdoge *Watchdoge) {
	client := redis.NewTCPClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
	})

	response, err := ssh_params.ConnectAndRun("cat /proc/" + watchdoge.Pid + "/status | grep " + watchdoge.Stat)
	if err != nil {	fmt.Println(err.Error()) }
	fmt.Println(strings.TrimSpace(response))
	b := strings.Split(response, " ")
	client.Set("metrics:" + ssh_params.Server + ":" + watchdoge.Pid + ":" + b[0] + ":" + string(int32(time.Now().Unix())), b[2]+b[3])
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
	w.Header().Set("Content-Type", "application/json")
	stream := ServerStream { w, r }

	server_name := r.URL.Query().Get("server")

	client := redis.NewTCPClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
	})

	server, _ := client.Get("servers:" + server_name + ":ip").Result()
	user, _ := client.Get("servers:" + server_name + ":ssh_user").Result()

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

	renderJSON(stream, Status{true})
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stream := ServerStream { w, r }
	renderJSON(stream, Status{true})
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

func ps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stream := ServerStream { w, r }

	server_name := r.URL.Query().Get("server")

	client := redis.NewTCPClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
	})

	server, err := client.Get("servers:" + server_name + ":ip").Result()
	user, err := client.Get("servers:" + server_name + ":ssh_user").Result()

	ssh_params := &easyssh.MakeConfig {
        User: user,
        Server: server,
        Key: "/.ssh/id_rsa",
    }

	response, err := ssh_params.ConnectAndRun("ps axho pid,command --sort rss")
	if err != nil {
		renderJSON(stream, Status{false})
		return
	} else {
		processes := ParseProcesses(response)
		renderJSON(stream, processes)
	}
}

func loadConfig() {
	buf, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Println("error: %v", err)
	}
	err = yaml.Unmarshal(buf, &servers)
	if err != nil {
		fmt.Println("error: %v", err)
	}
}

func setServers() {
	client := redis.NewTCPClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
	})

	for server_name, server_params := range servers.Servers {
		if err := client.Set("servers:" + server_name + ":ip", server_params[0]).Err(); err != nil {
	    	panic(err)
		}
		if err := client.Set("servers:" + server_name + ":ssh_user", server_params[1]).Err(); err != nil {
	    	panic(err)
		}
		fmt.Println("Server " + server_name + " published in redis store")
	}
}

func main() {
	port := "8080"
	loadConfig()
	setServers()
	http.HandleFunc("/status", status)
	http.HandleFunc("/api", api_handler)
	http.HandleFunc("/ps", ps)
	fmt.Println("Server running on port", port)
	http.ListenAndServe(":" + port, nil)
}