package main

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/satori/go.uuid"
	"math"
	"time"

	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"serv/messages"
)

type Monster struct {
	Position mgl32.Vec3
	UUID     uuid.UUID
}

func (m *Monster) nearestPlayer() *Connection {
	var connection *Connection
	dist := math.Inf(1)
	for _, con := range connections {
		curDist := float64(con.Position.Sub(m.Position).Len())
		if curDist < dist {
			dist = curDist
			connection = con
		}
	}
	return connection
}

func (m *Monster) Move() {
	dest := m.nearestPlayer()
	if dest != nil {
		dir := dest.Position.Sub(m.Position).Normalize().Mul(0.2)
		m.Position = m.Position.Add(dir)

		msg := messages.MonsterMoveBroadcastMsg{}
		msg.Type = messages.MonsterMoveBroadcast
		msg.Position = m.Position
		msg.UUID = m.UUID
		broadcast(msg, "")
	}
}

func NewMonster() *Monster {
	monster := Monster{}

	monster.UUID = uuid.Must(uuid.NewV4())
	return &monster
}

var connections map[string]*Connection
var monsters []*Monster
var level = 0
var pc net.PacketConn

type Parser interface {
	Parse([]byte) error
}

type Connection struct {
	addr     net.Addr
	uuid     uuid.UUID
	Position mgl32.Vec3
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
		c.Position = upmsg.Position
		broadcast(posBroadcast, c.addr.String())
	default:
		panic("Unkown msg type")
	}
}

func sendResponse(addr net.Addr, msg interface{}) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, msg)
	if err != nil {
		panic(err)
	}

	pc.WriteTo(buffer.Bytes(), addr)
}

func broadcast(msg interface{}, initiatorAddr string) {
	for addrString, con := range connections {
		if addrString != initiatorAddr {
			sendResponse(con.addr, msg)
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
	con = &Connection{addr: addr, uuid: uuid}
	connections[addrString] = con
	fmt.Printf("New connection: %s\n", uuid)

	return con
}

func spawner() {
	for {
		if len(monsters) == 0 {
			level++
			monsterNumber := level * level
			fmt.Printf("Level %d - spawning %d monsters\n", level, monsterNumber)
			for i := 0; i < monsterNumber; i++ {
				monster := NewMonster()
				monsters = append(monsters, monster)
			}
		}
	}
}

func mover() {
	for {
		time.Sleep(time.Millisecond * 75)
		for _, monster := range monsters {
			monster.Move()
		}
	}
}

func main() {
	connections = make(map[string]*Connection)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	go spawner()
	go mover()

	var err error
	pc, err = net.ListenPacket("udp", fmt.Sprintf(":%s", port))
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
