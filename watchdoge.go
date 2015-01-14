package main

import "flag"
import "fmt"
import "golang.org/x/crypto/ssh"
import "os/user"
import "io/ioutil"
import "bytes"

func RecordStats(procname string, subprocname string, period string, server string, config *ssh.ClientConfig) string {

	client, err := ssh.Dial("tcp", server+":22", config)
	if err != nil {
		panic("Failed to dial: "+ err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	command := "ps ax -o pid,command | grep " + procname + " | grep " + subprocname + " | grep -v bash"
	if err := session.Run(command); err != nil {
		panic("Failed to run: " + err.Error())
	}
	return b.String()
}

func getKeyFile() (pubkey ssh.Signer){

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

func main() {

	user := flag.String("user", "root", "SSH username")
	server := flag.String("server", "localhost", "Server name for watchdogging")
	procname := flag.String("procname","ssh","Process name")
	subprocname := flag.String("subprocname", "", "Subprocess name")
	period := flag.String("period", "1", "Period between statistics checks")

	flag.Parse()

	fmt.Println("Server:", *server)	
	fmt.Println("SSH username:", *user)
	fmt.Println("Process name:", *procname)
	fmt.Println("Subprocess name:", *subprocname)
	fmt.Println("Checks period:", *period)

	pubkey := getKeyFile()
	config := &ssh.ClientConfig{
			User: *user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(pubkey)},
		}
	stats := RecordStats(*procname, *subprocname, *period, *server, config)
	fmt.Println("\nProcess info:")
	fmt.Println(stats)
}
