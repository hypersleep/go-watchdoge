package test

import(
	"net/http"
	"strconv"
)

type Watchdoge struct {
	Procname, Subprocname string
	Period, Iterations int
}

type ServerStream struct {
	Write http.ResponseWriter
	Read *http.Request
}

func ParseParam (r *http.Request, param string) (parsed_param string) {
	parsed_param = r.URL.Query().Get(param)
	return
}

func ParseIntParam (r *http.Request, param string) (parsed_param int) {
	param = r.URL.Query().Get(param)
	parsed_param, _ = strconv.Atoi(param)
	return
}

func (watchdoge *Watchdoge) GetProcname (stream ServerStream) {
	procname := ParseParam(stream.Read, "procname")
	watchdoge.Procname = procname
}

func (watchdoge *Watchdoge) GetSubprocname (stream ServerStream) {
	subprocname := ParseParam(stream.Read, "subprocname")
	watchdoge.Subprocname = subprocname
}

func (watchdoge *Watchdoge) GetPeriod (stream ServerStream) {
	period := ParseIntParam(stream.Read, "period")
	watchdoge.Period = period
}

func (watchdoge *Watchdoge) GetIterations (stream ServerStream) {
	iterations := ParseIntParam(stream.Read, "iterations")
	watchdoge.Iterations = iterations
}