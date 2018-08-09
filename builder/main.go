package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/masterzen/winrm"
	"github.com/packer-community/winrmcp/winrmcp"
)

const (
	copyTimeout = 300 * time.Second
)

var (
	hostname = flag.String("hostname", "localhost", "Hostname of remote Windows server")
	username = flag.String("username", "ContainerAdministrator", "Username on remote Windows server")
	password = flag.String("password", "", "Password on remote Windows server")
)

func main() {
	// Parse flags for host, username, password
	flag.Parse()
	r := &Remote{
		hostname: hostname,
		username: username,
		password: password,
	}
	// Copy workspace to remote machine
	r.copy()

	// Execute on remote
	r.run("ipconfig /all")
}

// Remote represents a remote Windows server.
type Remote struct {
	hostname *string
	username *string
	password *string
}

func (r *Remote) copy() error {
	c, err := winrmcp.New(*r.hostname, &winrmcp.Config{
		Auth:                  winrmcp.Auth{User: *r.username, Password: *r.password},
		Https:                 false,
		Insecure:              true,
		TLSServerName:         "",
		CACertBytes:           nil,
		OperationTimeout:      copyTimeout,
		MaxOperationsPerShell: 15,
	})
	if err != nil {
		log.Printf("Error copying workspace to remote: %+v", err)
		return err
	}
	return c.Copy("/workspace", `C:\workspace`)
}

func (r *Remote) run(command string) error {
	stdin := bytes.NewBufferString(command)
	endpoint := winrm.NewEndpoint(*r.hostname, 5985, false, false, nil, nil, nil, 0)
	w, err := winrm.NewClient(endpoint, *r.username, *r.password)
	if err != nil {
		log.Printf("Error creating remote client: %+v", err)
		return err
	}
	shell, err := w.CreateShell()
	if err != nil {
		log.Printf("Error creating remote shell: %+v", err)
		return err
	}
	var cmd *winrm.Command
	cmd, err = shell.Execute("cmd.exe")
	if err != nil {
		log.Printf("Error executing remote command: %+v", err)
		return err
	}

	go io.Copy(cmd.Stdin, stdin)
	go io.Copy(os.Stdout, cmd.Stdout)
	go io.Copy(os.Stderr, cmd.Stderr)

	cmd.Wait()
	shell.Close()
	return nil
}
