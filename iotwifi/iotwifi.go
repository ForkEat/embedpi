// IoT Wifi packages is used to manage WiFi AP and Station (client) modes on
// a Raspberry Pi or other arm device. This code is intended to run in it's
// corresponding Alpine docker container.

package iotwifi

import (
	"bufio"
	"embed/config"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/bhoriuchi/go-bunyan/bunyan"
	"go.uber.org/zap"
)

// CmdRunner runs internal commands allows output handlers to be attached.
type CmdRunner struct {
	Messages chan CmdMessage
	Handlers map[string]func(CmdMessage)
	Commands map[string]*exec.Cmd
}

// CmdMessage structures command output.
type CmdMessage struct {
	Id      string
	Command string
	Message string
	Error   bool
	Cmd     *exec.Cmd
	Stdin   *io.WriteCloser
}

// RunWifi starts AP and Station modes.
func RunWifi(log bunyan.Logger, messages chan CmdMessage, setupCfg *config.SetupCfg) {

	zap.S().Info("Loading IoT Wifi...")

	cmdRunner := CmdRunner{
		Messages: messages,
		Handlers: make(map[string]func(cmsg CmdMessage), 0),
		Commands: make(map[string]*exec.Cmd, 0),
	}

	command := &Command{
		Runner:   cmdRunner,
		SetupCfg: setupCfg,
	}

	// listen to kill messages
	cmdRunner.HandleFunc("kill", func(cmsg CmdMessage) {
		zap.S().Error("GOT KILL")
		os.Exit(1)
	})

	wpacfg := NewWpaCfg(log, setupCfg)

	wpacfg.StartAP()

	time.Sleep(10 * time.Second)

	command.StartWpaSupplicant()

	// Scan
	time.Sleep(5 * time.Second)
	wpacfg.ScanNetworks()

	command.StartDnsmasq()

	// TODO: check to see if we are stuck in a scanning state before
	// if in a scanning state set a timeout before resetting
	go func() {
		for {
			wpacfg.ScanNetworks()
			time.Sleep(30 * time.Second)
		}
	}()

	// staticFields for logger
	staticFields := make(map[string]interface{})

	// command output loop (channel messages)
	// loop and log
	//
	for {
		out := <-messages // Block until we receive a message on the channel

		staticFields["cmd_id"] = out.Id
		staticFields["cmd"] = out.Command
		staticFields["is_error"] = out.Error

		zap.S().Info(staticFields, out.Message)

		if handler, ok := cmdRunner.Handlers[out.Id]; ok {
			handler(out)
		}
	}
}

// HandleFunc is a function that gets all channel messages for a command id
func (c *CmdRunner) HandleFunc(cmdId string, handler func(cmdMessage CmdMessage)) {
	c.Handlers[cmdId] = handler
}

// ProcessCmd processes an internal command.
func (c *CmdRunner) ProcessCmd(id string, cmd *exec.Cmd) {
	zap.S().Debug("ProcessCmd got %s", id)

	// add command to the commands map TODO close the readers
	c.Commands[id] = cmd

	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	cmdStderrReader, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	stdOutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for stdOutScanner.Scan() {
			c.Messages <- CmdMessage{
				Id:      id,
				Command: cmd.Path,
				Message: stdOutScanner.Text(),
				Error:   false,
				Cmd:     cmd,
			}
		}
	}()

	stdErrScanner := bufio.NewScanner(cmdStderrReader)
	go func() {
		for stdErrScanner.Scan() {
			c.Messages <- CmdMessage{
				Id:      id,
				Command: cmd.Path,
				Message: stdErrScanner.Text(),
				Error:   true,
				Cmd:     cmd,
			}
		}
	}()

	err = cmd.Start()

	if err != nil {
		panic(err)
	}
}
