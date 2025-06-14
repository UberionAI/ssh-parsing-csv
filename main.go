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
	//Structure from SSH settings
	type SettingsSSH struct {
		Username     string
		Password     string
		SudoPassword string
		Hostname     string
	}

	SSHpreset := SettingsSSH{
		Username:     os.Getenv("SSH_USERNAME"),
		Password:     os.Getenv("SSH_PASSWORD"),
		SudoPassword: os.Getenv("SSH_SUDO_PASSWORD"),
		Hostname:     os.Getenv("SSH_HOSTNAME"),
	}

	//Set flags
	commandsFlag := flag.String("commands", "", "Comma separated list of commands to run")
	nameFileFlag := flag.String("name", "output.txt", "Name for a file")

	//Parsing flags
	flag.Parse()
	if SSHpreset.Hostname == "" || SSHpreset.Password == "" || SSHpreset.SudoPassword == "" {
		log.Fatal("SSH_USERNAME or SSH_PASSWORD or SSH_SUDO_PASSWORD environment variables not set")
	}

	var commands []string
	if *commandsFlag != "" {
		initCommands := strings.Split(*commandsFlag, ",") //separate string commands with ","
		for _, cmd := range initCommands {
			cmd = strings.TrimSpace(cmd)
			prepareCommand := fmt.Sprintf("echo '%s' | sudo -S %s", SSHpreset.SudoPassword, cmd)
			commands = append(commands, prepareCommand)
		}
	}

	config := &ssh.ClientConfig{
		User: SSHpreset.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(SSHpreset.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//Connection stage
	client, err := ssh.Dial("tcp", SSHpreset.Hostname, config)
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}
	defer client.Close()

	//Name for saving file (default name is "output.txt")
	file, err := os.Create(*nameFileFlag)
	if err != nil {
		log.Fatal(err, "Error creating file.")
	}

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
		outputStr := fmt.Sprintf("____________________________\nInput command for host %s: %s\n____________________________\nTHE RESULT:\n%s\n", SSHpreset.Hostname, *commandsFlag, stdoutBuf.String())
		file, err = os.OpenFile(*nameFileFlag, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err, "Error appending to file.")
		}
		defer file.Close()

		//Formating new lines to .txt file
		_, err = fmt.Fprintf(file, "%s\n", outputStr)
		if err != nil {
			log.Fatal(err, "Error formating .txt file")
		}
	}
	fmt.Printf("Successfully created file %s!\n", *nameFileFlag)
}
