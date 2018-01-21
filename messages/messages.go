package messages

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/satori/go.uuid"
)

const (
	UpdatePosition = iota
	PositionBroadcast
	MonsterMoveBroadcast
)

type Message struct {
	Type uint8
}

type UpdateLocationMsg struct {
	Message
	Position mgl32.Vec3
	Rotation mgl32.Quat
}

type MonsterMoveBroadcastMsg struct {
	Message
	Position mgl32.Vec3
	UUID     uuid.UUID
}

type PositionBroadcastMsg struct {
	UpdateLocationMsg
	UUID uuid.UUID
}
