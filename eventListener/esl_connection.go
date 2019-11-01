package eventListener

import (
	"fmt"
	ESL "github.com/0x19/goesl"
	"log"
)

type ESLConnection struct {
	Connection *ESL.Client
	ch         chan *ESL.Message
	active     bool
}

func NewESLConnection(connection *ESL.Client, ch chan *ESL.Message) *ESLConnection {
	res := &ESLConnection{
		Connection: connection,
		ch:         ch,
		active:     false,
	}
	go res.run()
	return res
}

func (ec *ESLConnection) SubscribeEvent(eventName string) error {
	err := ec.Connection.Send(fmt.Sprintf("event json %s", eventName))
	return err
}

func (ec ESLConnection) IsActive() bool {
	return ec.active
}

func (ec *ESLConnection) run() {
	ec.active = true
	for {
		msg, err := ec.Connection.ReadMessage()
		if err != nil {
			log.Printf("Got error reading ESL message: %v", err)
			ec.active = false
			break
		}
		if msg == nil {
			continue
		}
		ec.ch <- msg
	}
	if err := ec.Connection.Close(); err != nil {
		log.Printf("Got error closing ESL connection: %v", err)
	}
}
