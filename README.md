# Go-WatchDoge

## Description

This app is created for watching a process on remote host through SSH connection.

Now you can watch only /proc/`pid`/status out, but many awesome features will be added coming soon...

## Config

First create a config file. 

config.yml

`
port: 8080
redis: localhost:6379
servers:
  digital-ocean109:
    - 10.200.82.109
    - hypersleep
  local135:
    - 172.16.1.1
    - hypersleep
`

Yep. This app uses awesome Redis NoSQL. Ensure Redis is avaliable.

## Usage

You can simply curl'ing it:

`curl localhost:8080/status`

Get processes list:

`curl localhost:8080/ps?server=digital-ocean109`

Copy pid of process and forward to next step.

Pull VmRSS metric 3 times from 12938 pid:

`curl http://localhost:8080/api?server=digital-ocean109\&pid=12938\&period=1\&iterations=3\&stat=VmRSS`

You can watch metrics in STDOUT output and your Redis store.

## Roadmap

* Add moitoring jobs (list, create, start, stop, remove)
* JS web frontend
