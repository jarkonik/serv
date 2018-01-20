package messages

const (
	UpdatePosition = iota
	PositionBroadcast
)

type Message struct {
	Type uint8
}

type UpdateLocationMsg struct {
	Message
	X float32
	Y float32
	Z float32
}

type PositionBroadcastMsg struct {
	UpdateLocationMsg
}
