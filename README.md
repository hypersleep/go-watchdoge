# Go-WatchDoge

## Description

This app is created for watching a process on remote host through SSH connection.

Now you can watch only pidstat utilty out, but many awesome features will be added coming soon...

App working in server mode and listen *8080* port.

You can simply curl'ing it:

`curl localhost:8080/status`

`curl http://localhost:8080/api?user=v.spirenkov\&server=10.200.82.109\&procname=istream3\&subprocname=cdn\&period=1\&iterations=3`

## Roadmap

* Add jobs (list, create, start, stop, remove)
* Serving jobs and metrics in Redis
* Web frontend
