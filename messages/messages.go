package messages

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/satori/go.uuid"
)

const (
	UpdatePosition = iota
	PositionBroadcast
)

type Message struct {
	Type uint8
}

type UpdateLocationMsg struct {
	Message
	Position mgl32.Vec3
	Rotation mgl32.Quat
}

type PositionBroadcastMsg struct {
	UpdateLocationMsg
	UUID uuid.UUID
}
