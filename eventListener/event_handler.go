package eventListener

import (
	ESL "github.com/0x19/goesl"
)

type Handler func(event *ESL.Message)

type EventHandler struct {
	EventName string
	Handle    Handler
}
