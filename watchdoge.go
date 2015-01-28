package main

import(
	"fmt"
	"strings"
	"strconv"
	"time"
	"github.com/hypersleep/easyssh"
)

type Watchdoge struct {
	Pid, Stat string 
	Period, Iterations int
}

func RunPidstat(ssh_params *easyssh.MakeConfig, watchdoge *Watchdoge) {
	response, err := ssh_params.ConnectAndRun("cat /proc/" + watchdoge.Pid + "/status | grep " + watchdoge.Stat)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(strings.TrimSpace(response))
		b := strings.Split(response, " ")
		key := "metrics:" + ssh_params.Server
		key += ":" + watchdoge.Pid
		key += ":" + strings.TrimSpace(b[0]) + strconv.Itoa(int(time.Now().Unix()))
		value := b[2] + b[3]
		redis_client.Set(key, value)
	}	
	return
}

func PullRemoteProcessMetrics(ssh_params *easyssh.MakeConfig, watchdoge *Watchdoge) {
	fmt.Println("Process info:")
	for i := 0; i < watchdoge.Iterations; i++ {
		go RunPidstat(ssh_params, watchdoge)
		time.Sleep(time.Duration(watchdoge.Period) * time.Second)
	}
}
