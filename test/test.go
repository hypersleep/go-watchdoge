package test

import(
	"fmt"
	"net/http"
	"golang.org/x/crypto/ssh"
	"errors"
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

type EasySSH struct {
	Server string
	Config *ssh.ClientConfig
}

func ParseParam (r *http.Request, param string) (parsed_param string, err error) {
	parsed_param = r.URL.Query().Get(param)
	if len(parsed_param) == 0 {
		err = errors.New("Ensure " + param)
		return
	}
	return
}

func ParseIntParam (r *http.Request, param string) (parsed_param int, err error) {
	param = r.URL.Query().Get(param)
	parsed_param, _ = strconv.Atoi(param)
	return
}

func (doge Watchdoge) GetParams (stream ServerStream) (params Watchdoge, err error) {
	procname, err := ParseParam(stream.Read, "procname")
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to fetch procname parameter!")
		return
	}
	subprocname, err := ParseParam(stream.Read, "subprocname")
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to fetch subprocname parameter!")
		return
	}
	period, err := ParseIntParam(stream.Read, "period")
	if err != nil {
		period = 1
	}
	iterations, err := ParseIntParam(stream.Read, "iterations")
	if err != nil {
		iterations = 3
	}

	params = doge {procname, subprocname, period, iterations}
	return
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

func (conf EasySSH) GetConfig (stream ServerStream) (easy_config EasySSH, err error) {
	user, err := ParseParam(stream.Read, "user")
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to fetch user parameter!")
		return
	}
	server, err := ParseParam(stream.Read, "server")
	if err != nil {
		fmt.Fprintf(stream.Write, "Failed to fetch server parameter!")
		return
	}
	pubkey := getKeyFile()
	config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}

	easy_config = conf {server, config}
	return
}