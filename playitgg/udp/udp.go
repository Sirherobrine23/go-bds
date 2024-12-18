/*
Package to process clients from UDP Listener
*/
package udp

import (
	"bytes"
	"net"
	"sync"
	"time"
)

var (
	_ net.Listener = &UDP{}
	_ net.Conn     = &ClientStatus{}
)

type ClientStatus struct {
	net.Conn
	ConnWrite  net.Conn
	LastUpdate time.Time
}

type UDP struct {
	*net.UDPConn

	clientsLocker    sync.Mutex
	clients          map[string]*ClientStatus
	connectionsErr   chan error
	clientConnection chan net.Conn
}

func Listen(network string, laddr *net.UDPAddr) (*UDP, error) {
	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		return nil, err
	}

	udpConn := &UDP{conn, sync.Mutex{}, make(map[string]*ClientStatus), make(chan error), make(chan net.Conn)}
	go udpConn.process()
	return udpConn, nil
}

func (udp *UDP) Addr() net.Addr { return udp.UDPConn.LocalAddr() }

func (udp *UDP) process() {
	bufferReader := make([]byte, 32*1024)
	for {
		bytesRead, from, err := udp.UDPConn.ReadFromUDP(bufferReader)
		if err != nil {
			udp.connectionsErr <- err
			return
		}

		udp.clientsLocker.Lock()
		// Process current or delete client
		if status, ok := udp.clients[from.String()]; ok {
			if status.LastUpdate.Compare(time.Now()) > 80_000_000 {
				go status.ConnWrite.Close()
				go status.Conn.Close()
				delete(udp.clients, from.String())
			} else {
				status.LastUpdate = time.Now()
				cloneBuff := bytes.Clone(bufferReader[:bytesRead])
				go status.ConnWrite.Write(cloneBuff)
				status.LastUpdate = time.Now()
				udp.clientsLocker.Unlock() // Unlocker
				continue
			}
		}

		// Add new client to struct
		newClient := &ClientStatus{}
		newClient.Conn, newClient.ConnWrite = net.Pipe()
		newClient.LastUpdate = time.Now()
		udp.clients[from.String()] = newClient // Set new client
		go func(){
			buff := make([]byte, 32*1024)
			n, err := newClient.ConnWrite.Read(buff)
			if err != nil {
				return
			} else if _, err := udp.UDPConn.WriteToUDP(buff[:n], from); err != nil {
				return
			}
			newClient.LastUpdate = time.Now()
		}()
		udp.clientsLocker.Unlock() // unlock
	}
}

// Accept connections from listen UDP
func (udp *UDP) Accept() (net.Conn, error) {
	select {
	case err := <-udp.connectionsErr:
		return nil, err
	case conn := <-udp.clientConnection:
		return conn, nil
	}
}
