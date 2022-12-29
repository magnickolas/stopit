package stopit

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/phayes/freeport"
)

type StopItServer struct {
	Port     int
	StartNow bool
	Cmd      *exec.Cmd
}

func StopItServerWithFreePort(cmd *exec.Cmd, startNow bool) (StopItServer, error) {
	freePort, err := freeport.GetFreePort()
	if err != nil {
		return StopItServer{}, err
	}
	return StopItServer{
		Port:     freePort,
		StartNow: startNow,
		Cmd:      cmd,
	}, nil
}

func (self StopItServer) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", self.Port))
	if err != nil {
		panic(err)
	}
	run := make(chan any, 1)
	pause := make(chan any)
	paused := make(chan any, 1)
	if self.StartNow {
		go runCommandUntilPaused(pause, paused, *self.Cmd)
		run <- nil
	} else {
		paused <- nil
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn, run, pause, paused, self.Cmd)
	}
}

func handleConn(
	conn net.Conn,
	run chan any, pause chan any, paused chan any,
	cmd *exec.Cmd,
) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	byte, err := reader.ReadByte()
	if err != nil {
		panic(err)
	}
	if byte == 'r' {
		select {
		case <-paused:
			go runCommandUntilPaused(pause, paused, *cmd)
			run <- nil
		case <-run:
			run <- nil
		}
	} else if byte == 's' {
		select {
		case <-run:
			pause <- nil
		case <-pause:
			pause <- nil
		case <-paused:
			paused <- nil
		}
	}
}

func runCommandUntilPaused(pause chan any, paused chan any, cmd exec.Cmd) {
	if err := cmd.Start(); err != nil {
		log.Panic("failed to start process", cmd)
	}
	<-pause
	if err := cmd.Process.Kill(); err != nil {
		log.Panic("failed to kill process", cmd.Process.Pid)
	}
	paused <- nil
}
