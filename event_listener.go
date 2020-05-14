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
	"sync"
	"time"
)

type EventListener struct {
	// AMQPQueuesPool    []AMQPConnection
	ESLConnectionPool []*ESLConnection
	EventHandlers     []*EventHandler
	events            chan *ESL.Message
	evListMutex       sync.Mutex
	eslConnListMutex  sync.Mutex
	stop              bool
}

func NewEventListener() *EventListener {
	el := EventListener{
		// AMQPQueuesPool:    make([]AMQPConnection, 0),
		ESLConnectionPool: make([]*ESLConnection, 0),
		EventHandlers:     make([]*EventHandler, 0),
		events:            make(chan *ESL.Message),
		stop:              false,
	}
	go el.run()
	return &el
}

func (el *EventListener) OpenESLConnection(host, password string, port uint, timeout int) error {
	client, err := ESL.NewClient(host, port, password, timeout)
	if err != nil {
		return err
	}
	go client.Handle()
	eslConn := NewESLConnection(client, el.events)
	el.eslConnListMutex.Lock()
	el.ESLConnectionPool = append(el.ESLConnectionPool, eslConn)
	el.eslConnListMutex.Unlock()
	el.evListMutex.Lock()
	for i := range el.EventHandlers {
		go func() {
			if err := eslConn.SubscribeEvent(el.EventHandlers[i].EventName); err != nil {
				// todo add error handling
			}
		}()
	}
	el.evListMutex.Unlock()
	return nil
}

func (el *EventListener) AddEventHandler(eventName string, handler Handler) []error {
	errs := make([]error, 0)
	el.eslConnListMutex.Lock()
	for i := range el.ESLConnectionPool {
		go func() {
			if err := el.ESLConnectionPool[i].SubscribeEvent(eventName); err != nil {
				errs = append(errs, err)
			}
		}()
	}
	el.eslConnListMutex.Unlock()
	el.evListMutex.Lock()
	defer el.evListMutex.Unlock()
	el.EventHandlers = append(el.EventHandlers, &EventHandler{eventName, handler})
	if len(errs) > 0 {
		return errs
	}
	return nil
}

/*
func (el *EventListener) CloseESLConnection(connection ESLConnection) error {
	// todo need implementation
	return nil
}

func (el *EventListener) SubscribeAMQP() error {
	return nil
}

func (el *EventListener) UnsubscribeAMQP(connection AMQPConnection) error {
	// todo need implementation
	return nil
}

func (el *EventListener) RemoveEventHandler(handler Handler) error {
	// todo need implementation
	return nil
}
*/

func (el *EventListener) run() {
	for !el.stop {
		select {
		case msg := <-el.events:
			el.evListMutex.Lock()
			for i := range el.EventHandlers {
				event := msg.GetHeader("Event-Name")
				if event == "CUSTOM" {
					event = fmt.Sprintf("CUSTOM %s", msg.GetHeader("Event-Subclass"))
				}
				if event != el.EventHandlers[i].EventName {
					continue
				}
				go el.EventHandlers[i].Handle(msg)
			}
			el.evListMutex.Unlock()
		default:
			time.Sleep(1000)
		}
	}
}
