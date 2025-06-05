package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		//fmt.Printf("Input command for host %s: %s\n____________________________\nTHE RESULT:\n%s\n", hostname, *commandsFlag, stdoutBuf.String())

		csvBuf := bytes.NewBuffer(stdoutBuf.Bytes())
		err = saveStdoutBuf("csv_output", csvBuf)
		if err != nil {
			log.Printf("Error saving stdout from command: %v", err)
		}
		continue
	}
	log.Println("Successfully saved all output into a csv!")
}

// Function for saving results into a csv file
func saveStdoutBuf(dir string, stdoutBuf *bytes.Buffer) error {
	if len(stdoutBuf.Bytes()) == 0 {
		fmt.Errorf("stdoutbuf is empty, file will not be created!")
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Errorf("Error creating directory %s: %v", dir, err)
	}

	timestamp := time.Now().Format("20060102150405")
	filename := filepath.Join(dir, "ResultOut_"+timestamp+".csv")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()

	//csv-writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	//bytes to string

	txt := stdoutBuf.String()
	if err := writer.Write([]string{txt}); err != nil {
		fmt.Errorf("Error writing to file: %v", err)
	}
	return err
}
