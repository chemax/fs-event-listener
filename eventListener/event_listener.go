package eventListener

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
