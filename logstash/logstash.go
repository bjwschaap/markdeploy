package logstash

import (
	"net"
	"fmt"
	"time"
	"errors"
)

type Logstash struct {
	Hostname string
	Port int
	Connection *net.TCPConn
	Timeout int
}

func New(hostname string, port int, timeout int) *Logstash {
	return &Logstash{
		Hostname: hostname,
		Port: port,
		Connection: nil,
		Timeout: timeout,
	}
}

func (l *Logstash) Dump() {
	fmt.Println("Hostname:   ", l.Hostname)
	fmt.Println("Port:       ", l.Port)
	fmt.Println("Connection: ", l.Connection)
	fmt.Println("Timeout:    ", l.Timeout)
}

func (l *Logstash) SetTimeouts() {
	deadline := time.Now().Add(time.Duration(l.Timeout) * time.Millisecond)
	l.Connection.SetDeadline(deadline)
	l.Connection.SetWriteDeadline(deadline)
	l.Connection.SetReadDeadline(deadline)
}

func (l *Logstash) Connect() (*net.TCPConn, error) {
	var connection *net.TCPConn
	service := fmt.Sprintf("%s:%d", l.Hostname, l.Port)
	addr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		return connection, err
	}
	connection, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return connection, err
	}
	if connection != nil {
		l.Connection = connection
		l.Connection.SetLinger(0) // default -1
		l.Connection.SetNoDelay(true)
		l.Connection.SetKeepAlive(true)
		l.Connection.SetKeepAlivePeriod(time.Duration(5) * time.Second)
		l.SetTimeouts()
	}
	return connection, err
}

func (l *Logstash) Writeln(message []byte) (error) {
	var err = errors.New("tcp connection is nil")
	message = append(message, '\n')
	if l.Connection != nil {
		_, err = l.Connection.Write(message)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				l.Connection.Close()
				l.Connection = nil
				if err != nil {
					return err
				}
			} else {
				l.Connection.Close()
				l.Connection = nil
				return err
			}
		} else {
			// Successful write! Let's extend the timeout.
			l.SetTimeouts()
			return nil
		}
	}
	return err
}
