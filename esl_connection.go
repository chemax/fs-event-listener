/*
Copyright (c) 2019 Dmitrii Borisov <dborisov@mail.ru>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit
persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
package fsEventListener

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
