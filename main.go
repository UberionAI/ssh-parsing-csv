package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
)

func main() {

	//Getting server connection variables from file .env and flags
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	sshUsername := os.Getenv("SSH_USERNAME")
	sshPassword := os.Getenv("SSH_PASSWORD")
	sshSudoPassword := os.Getenv("SSH_SUDO_PASSWORD")
	hostname := os.Getenv("SSH_HOSTNAME")

	//Set flags
	commandsFlag := flag.String("commands", "", "Comma separated list of commands to run")

	//Parsing flags
	flag.Parse()

	if sshUsername == "" || sshPassword == "" || sshSudoPassword == "" {
		log.Fatal("SSH_USERNAME or SSH_PASSWORD or SSH_SUDO_PASSWORD environment variables not set")
	}

	var commands []string
	if *commandsFlag != "" {
		initCommands := strings.Split(*commandsFlag, ",") //separate string commands with ","
		for _, cmd := range initCommands {
			cmd = strings.TrimSpace(cmd)
			prepareCommand := fmt.Sprintf("echo '%s' | sudo -S %s", sshSudoPassword, cmd)
			commands = append(commands, prepareCommand)
		}
	}

	config := &ssh.ClientConfig{
		User: sshUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//Connection stage
	client, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}
	defer client.Close()

	for _, cmd := range commands {
		session, err := client.NewSession()
		if err != nil {
			log.Fatalf("Error creating session: %v", err)
			continue
		}
		defer session.Close()

		var stdoutBuf, stderrBuf bytes.Buffer
		session.Stdout = &stdoutBuf
		session.Stderr = &stderrBuf

		if err := session.Run(cmd); err != nil {
			//log.Printf("Error running command: %s: %v, stderr: %s", cmd, err, stderrBuf.String())
			log.Printf("Error running command: %s, stderr: %s", *commandsFlag, stderrBuf.String())
			continue
		}

		//Resulting output including commands and errors
		fmt.Printf("Input command for host %s: %s\n____________________________\nTHE RESULT:\n%s\n", hostname, *commandsFlag, stdoutBuf.String())
	}
}
