package stopit

import (
	"bufio"
	"fmt"
	"net"
)

type StopIt struct {
	Port int
}

type action byte

var (
	run  action = 'r'
	stop action = 's'
)

func (self StopIt) Run() error {
	return self.perform(run)
}

func (self StopIt) Stop() error {
	return self.perform(stop)
}

func (self StopIt) perform(t action) error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", self.Port))
	defer conn.Close()
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(conn)
	if err := writer.WriteByte(byte(t)); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
