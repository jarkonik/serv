package main

import (
	"bytes"
	"github.com/satori/go.uuid"

	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"serv/messages"
)

var connections map[string]*Connection

type Parser interface {
	Parse([]byte) error
}

type Connection struct {
	addr net.Addr
	pc   net.PacketConn
	uuid uuid.UUID
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
	case messages.UpdatePosition:
		upmsg := messages.UpdateLocationMsg{}
		err := binary.Read(rdr, binary.LittleEndian, &upmsg)
		if err != nil {
			panic(err)
		}

		posBroadcast := messages.PositionBroadcastMsg{}
		posBroadcast.Type = messages.PositionBroadcast
		posBroadcast.Position = upmsg.Position
		posBroadcast.Rotation = upmsg.Rotation
		posBroadcast.UUID = c.uuid
		broadcast(c.pc, posBroadcast, c.addr.String())
	default:
		panic("Unkown msg type")
	}
}

func sendResponse(conn net.PacketConn, addr net.Addr, msg interface{}) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, msg)
	if err != nil {
		panic(err)
	}

	conn.WriteTo(buffer.Bytes(), addr)

}

func broadcast(conn net.PacketConn, msg interface{}, initiatorAddr string) {
	for addrString, con := range connections {
		if addrString != initiatorAddr {
			sendResponse(conn, con.addr, msg)
		}
	}
}

func findOrCreateConnection(addr net.Addr, pc net.PacketConn) *Connection {
	addrString := addr.String()

	con := connections[addrString]
	if con != nil {
		return con
	}

	uuid := uuid.Must(uuid.NewV4())
	con = &Connection{addr: addr, pc: pc, uuid: uuid}
	connections[addrString] = con
	fmt.Printf("New connection: %s\n", uuid)

	return con
}

func main() {
	connections = make(map[string]*Connection)

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
		con := findOrCreateConnection(addr, pc)
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
