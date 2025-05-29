package main

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

func main() {

	//Getting server connection variables from file .env
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	sshUsername := os.Getenv("SSH_USERNAME")
	sshPassword := os.Getenv("SSH_PASSWORD")
	sshSudoPassword := os.Getenv("SSH_SUDO_PASSWORD")
	hostname := os.Getenv("HOSTNAME")

	if sshUsername == "" || sshPassword == "" || sshSudoPassword == "" {
		log.Fatal("SSH_USERNAME or SSH_PASSWORD or SSH_SUDO_PASSWORD environment variables not set")
	}
	//fmt.Println(sshUsername, sshPassword, sshSudoPassword)

	config := &ssh.ClientConfig{
		User: sshUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}
	defer client.Close()

	//Simple one command for host
	command := []string{
		fmt.Sprintf("echo '%s' | sudo -S iptables -L -n -v --line-numbers", sshSudoPassword),
	}

	//New session is creating
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}
	defer session.Close()
	for _, cmd := range command {
		var stdoutBuf, stderrBuf bytes.Buffer
		session.Stdout = &stdoutBuf
		session.Stderr = &stderrBuf
		err = session.Run(cmd)
		if err != nil {
			log.Fatalf("Error running command: %s, \n%v, \n%s", cmd, err, stderrBuf)
		}
		fmt.Println("The result: \n", string(stdoutBuf.Bytes()))
	}
}
