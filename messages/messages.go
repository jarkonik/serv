package messages

type Message struct {
	Type uint8
}

type UpdateLocationMsg struct {
	Message
	X uint8
	Y uint8
	Z uint8
}
