package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
)

// Slice with list of VM ip
var ipList []string

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
	if SSHpreset.Username == "" || SSHpreset.Password == "" || SSHpreset.SudoPassword == "" {
		log.Fatal("SSH_USERNAME or SSH_PASSWORD or SSH_SUDO_PASSWORD environment variables not set")
	}

	if _, err := os.Stat("VM_ListSSH.txt"); err == nil {
		file, err := os.Open("VM_ListSSH.txt")
		if err != nil {
			log.Fatal(err, "Can't read a file VM_ListSSH.txt with hostnames ")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ip := strings.TrimSpace(scanner.Text())
			if ip != "" {
				ipList = append(ipList, ip)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err, "Can't read a file VM_ListSSH.txt")
		}
		fmt.Println("Scanned Ip addresses from .txt file are below:")
		for _, ip := range ipList {
			fmt.Printf("%s\n", ip)
		}
		//another case - using the environment hostname
	} else if os.IsNotExist(err) {
		if SSHpreset.Hostname == "" {
			log.Fatal("Hostname not found in .env")
		}
		ipList = append(ipList, SSHpreset.Hostname)
		fmt.Printf("Hostname %s found in .env file\n", SSHpreset.Hostname)
	} else {
		log.Fatal(err, "error checking a file VM_ListSSH.txt: ", err)
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

	file, err := os.OpenFile(*nameFileFlag, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error creating or opening file: ", err)
	}
	defer file.Close()

	//Checking cycle to parse all the ip addresses from .txt file
	for _, ipHost := range ipList {
		addr := fmt.Sprintf("%s:22", ipHost)
		//Connection stage
		client, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			log.Fatalf("Error creating connection with %s: %v", ipHost, err)
			continue
		}
		defer client.Close()

		for _, cmd := range commands {
			session, err := client.NewSession()
			if err != nil {
				log.Fatalf("Error creating session for %s: %v", ipHost, err)
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
			outputStr := fmt.Sprintf("____________________________\nInput command for host %s: %s\n____________________________\nTHE RESULT:\n%s\n", ipHost, cmd, stdoutBuf.String())

			//Formating new lines to .txt file
			if _, err := fmt.Fprintf(file, "%s\n", outputStr); err != nil {
				log.Fatal("Error formating and writing the output into for hostname %s .txt file: %v", ipHost, err)
			}
		}
	}
	fmt.Printf("Successfully created file %s!\n", *nameFileFlag)
}
