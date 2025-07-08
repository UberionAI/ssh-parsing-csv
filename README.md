# Welcome to ssh-parsing-txt!

This Go tool lets you run any console commands on virtual machines (VMs) via SSH and save their output to a text file. Want to collect metrics about CPU, RAM, or disks, or check network configurations like routing tables or firewall rules? This project makes it quick and easy! It’s perfect for system administrators who need to automate tasks on servers.

## What’s it for?
Use **ssh-parsing-txt** at work to run commands on VMs, such as:
- Collecting metrics: CPU, RAM, and disk usage (e.g., `lsblk`, `df -h`).
- Checking network configurations: viewing routes, firewall rules, or network interfaces (e.g., `ip route`, `iptables -L -n -v`).  
The output of all commands is saved to a file, making it great for analysis, reporting, or debugging.

**Important**: Only commands that produce console output are supported (e.g., interactive commands like `top` or `nano` won’t work).

## What you’ll need
- Go installed (version 1.16 or later).
- SSH access to VMs with credentials (username, password, sudo password).
- A terminal and basic command-line skills.
- A text editor for configuring files.

## Step 1: Deployment
Here’s how to set up the project:

1. **Install Go**:
   - Download Go
   - Verify the installation:
     
     ```
     go version
     ```
     You should see a version, e.g., `go1.20.4`.

2. **Clone the repository**:
   ```
   git clone https://github.com/UberionAI/ssh-parsing-txt.git
   
   cd ssh-parsing-txt
   ```

3. **Install dependencies**:
   Run the following commands to fetch required libraries:
   ```
   go get github.com/joho/godotenv
   
   go get golang.org/x/crypto/ssh
   ```

4. **Set up VM access**:
   - Ensure you have SSH access to your VMs. Test it:
     ```
     
     ssh your_username@your_vm_ip
     
     ```
   - Verify that sudo is configured on the VMs (the `sudo -S` command should work with a password).

5. **Check your environment**:
   - Create a test `.env` file (see Step 2) and `commands.txt` (see Step 3).
   - Ensure `VM_ListSSH.txt` and `commands.txt` are readable.

## Step 2: Configure the setup
Create a `.env` file in the project root with your SSH credentials:
```
SSH_USERNAME=your_username
SSH_PASSWORD=your_password
SSH_SUDO_PASSWORD=your_sudo_password
SSH_HOSTNAME=your_VM_IP  (!!should be used only in one VM hostname case; otherwise just comment this line!!)
```
If you’re working with multiple VMs, create a `VM_ListSSH.txt` file, for example:
```
192.168.1.10
192.168.1.11
192.168.1.12
```

**Tip**: Ensure `.env` and `VM_ListSSH.txt` are not committed to Git (they’re already in `.gitignore`).

## Step 3: Prepare commands
****Either "commands.txt", or **setup** flag "-commands=..."****

*My advice to choice flag's way for a simlple one command, commands.txt otherwise.*
Create a `commands.txt` file with the commands you want to run. For example:
```
printf "CPU: $(nproc) threads\nRAM: $(free -h | awk '/Mem:/{print \"total=\" $2 \", used=\" $3}')\nDisks:\n$(lsblk -o NAME,SIZE | grep -v \"loop\")";ip route;iptables -L

```
These commands collect metrics (CPU, RAM, disks) and check network configurations (routes and firewall rules).

## Step 4: Run the tool
Run the program to execute the commands:
```
go run main.go -command-file="commands.txt" -name="output.txt"
```
Or specify commands directly:
```
go run main.go -commands="ip route" -name=network_info.txt
```

## Step 5: Check the results
Open the `output.txt` file — it will contain the results for each VM, for example:
```
____________________________
Input command for host 192.168.1.10: printf "CPU: $(nproc) threads\nRAM: $(free -h | awk '/Mem:/{print \"total=\" $2 \", used=\" $3}')\nDisks:\n$(lsblk -o NAME,SIZE | grep -v \"loop\")"
____________________________
THE RESULT:
CPU: 4 threads
RAM: total=7.8G, used=2.1G
Disks:
sda  50G
sdb 100G
____________________________
Input command for host 192.168.1.10: ip route
____________________________
THE RESULT:
default via 192.168.1.1 dev eth0
192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.10
____________________________
Input command for host 192.168.1.10: iptables -L
____________________________
THE RESULT:
Chain INPUT (policy ACCEPT)
target     prot opt source               destination
...
```

## Tips
- Verify that VMs are accessible via SSH before running.
- Ensure commands in `commands.txt` produce console output (interactive commands like `top` are not supported).
- If a command requires sudo, confirm that `SSH_SUDO_PASSWORD` is correct.
- For complex commands (with pipes or scripts), test them in a terminal first.
- The tool runs commands sequentially, so it may take time for many VMs.

## Security
Storing passwords in `.env` is not very secure. For production, consider using SSH keys:
1. Generate an SSH key: `ssh-keygen -t rsa`.
2. Copy the key to the VM: `ssh-copy-id your_username@your_vm_ip`.
3. Update the code to use keys instead of passwords (see the `golang.org/x/crypto/ssh` documentation).

## Possible improvements
- Add parallel command execution for faster processing.
- Support output formatting (e.g., JSON or CSV).
- Integrate with monitoring or log aggregation systems.

## Having issues?
If something doesn’t work:
- Check the terminal logs for connection or command errors.
- Ensure `.env` and `VM_ListSSH.txt` are correctly configured.
- Verify that commands produce console output.
- Create an issue on GitHub — we’ll figure it out together!

## Enjoy!
Now you can easily run any console commands on your VMs and save the results to a file. If you like the project, give it a star on GitHub!
