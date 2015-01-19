# Go-WatchDoge

## Description

This app is created for watching a process on remote host through SSH connection.

Now you can watch only /proc/<pid>/status out, but many awesome features will be added coming soon...

App working in server mode and listen *8080* port.

You can simply curl'ing it:

`curl localhost:8080/status`

Get processes list:

`curl localhost:8080/ps?user=root\&server=192.168.1.1`

Copy pid of process and forward to next step.

Pull VmRSS metric 3 times from 12938 pid:

`curl http://localhost:8080/api?user=root\&server=192.168.1.1\&pid=12938\&period=1\&iterations=3\&stat=VmRSS`

You can watch metrics in STDOUT output.

## Roadmap

* Add jobs (list, create, start, stop, remove)
* Serving jobs and metrics in Redis
* JS web frontend
