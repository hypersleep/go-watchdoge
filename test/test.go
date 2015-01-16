package test

import(
	"net/http"
	"golang.org/x/crypto/ssh"
	"strconv"
	"os/user"
	"io/ioutil"
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

type EasySSH struct {
	Server string
	Config *ssh.ClientConfig
}

func getKeyFile () (pubkey ssh.Signer){
	usr, err := user.Current()
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	pubkey, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}
	return
}

func (ssh_params *EasySSH) GetUser (stream ServerStream) {
	ssh_params.Server = ParseParam(stream.Read, "server")
}

func (ssh_params *EasySSH) GetConfig (stream ServerStream) {
	user := ParseParam(stream.Read, "user")
	pubkey := getKeyFile()
	ssh_params.Config = &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}
}