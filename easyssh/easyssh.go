package easyssh

import(
	"fmt"
	"strings"
	"net/http"
	"golang.org/x/crypto/ssh"
	"os/user"
	"io/ioutil"
	"bytes"
)

type EasySSH struct {
	Server string
	Config *ssh.ClientConfig
}

type ServerStream struct {
	Write http.ResponseWriter
	Read *http.Request
}

func ParseParam (r *http.Request, param string) (parsed_param string) {
	parsed_param = r.URL.Query().Get(param)
	return
}

func (ssh_params *EasySSH) ConnectAndRun (cmd string) (response string, err error) {
	client, err := ssh.Dial("tcp", ssh_params.Server+":22", ssh_params.Config)
	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
		return
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b	
	if err := session.Run(cmd); err != nil {
		panic("Failed to run: " + err.Error())
	}
	response = b.String()
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

func (ssh_params *EasySSH) CheckSSHConnection() {
	response, _ := ssh_params.ConnectAndRun("uname")
	response = strings.TrimSpace(response)
	fmt.Println(response)
}