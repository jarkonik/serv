package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"serv/messages"
)

var connections map[net.Addr]*Connection

const (
	Handshake = iota
	UpdatePosition
)

type Parser interface {
	Parse([]byte) error
}

type Connection struct {
	addr net.Addr
	X    uint8
	Y    uint8
	Z    uint8
}

func (c *Connection) Incoming(buffer []byte) {
	msg := messages.Message{}
	rdr := bytes.NewReader(buffer)

	err := binary.Read(rdr, binary.LittleEndian, &msg)
	if err != nil {
		panic(err)
	}
	rdr.Seek(0, 0)

	switch msg.Type {
	case UpdatePosition:
		upmsg := messages.UpdateLocationMsg{}
		err := binary.Read(rdr, binary.LittleEndian, &upmsg)
		if err != nil {
			panic(err)
		}
		c.X = upmsg.X
		c.Y = upmsg.Y
		c.Z = upmsg.Z
		fmt.Println(c)
	default:
		panic("Unkown msg type")
	}
}

func findOrCreateConnection(addr net.Addr) *Connection {
	con := connections[addr]
	if con != nil {
		return con
	}

	con = &Connection{addr: addr}
	return con
}

func main() {
	connections = make(map[net.Addr]*Connection)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%s", port))
	fmt.Printf("Server running on port %s\n", port)
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		buffer := make([]byte, 1024)
		var addr net.Addr
		_, addr, err := pc.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}
		con := findOrCreateConnection(addr)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Con error")
				}
			}()
			con.Incoming(buffer)
		}()
	}
}
