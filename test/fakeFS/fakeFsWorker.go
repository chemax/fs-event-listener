package fakeFS

import (
	"fmt"
	UUID "github.com/google/uuid"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	FsAuthAcceptedReply               = "Content-Type: command/reply\nReply-Text: +OK accepted\n\n"
	FsAuthDeniedReply                 = "Content-Type: command/reply\nReply-Text: -ERR invalid\n\n"
	FsExitReply                       = "Content-Type: command/reply\nReply-Text: +OK bye\n\n"
	FsErrCommandNotFound              = "Content-Type: command/reply\\rnReply-Text: -ERR command not found\n\n"
	FsDisconnectNotice                = "Content-Type: text/disconnect-notice\nContent-Length: 67\n\n"
	FsAuthInvite                      = "Content-Type: auth/request\n\n"
	FsPlainEventMessageHeaderTemplate = "Content-Length: %d\nContent-Type: text/event-plain\n\n"
	FsJsonEventMessageHeaderTemplate  = "Content-Length: %d\nContent-Type: text/event-json\n\n"
	BufLen                            = 4096
	SerializePlain                    = 0
	SerializeJson                     = 1
)

type Worker struct {
	stop         bool
	conn         net.Conn
	pass         string
	uuid         UUID.UUID
	events       []string
	customEvents []string
	evListsMutex sync.Mutex
	eventsChan   chan *Event
	serialize    int
}

func NewWorker(conn net.Conn, pass, uuid string, events chan *Event) *Worker {
	bytesUuid := []byte(strings.Replace(uuid, "-", "", -1))
	var resUuid UUID.UUID
	if tryUuid, err := UUID.FromBytes(bytesUuid); err != nil {
		resUuid, _ = UUID.NewUUID()
	} else {
		resUuid = tryUuid
	}
	return &Worker{
		conn:         conn,
		pass:         pass,
		events:       make([]string, 0),
		customEvents: make([]string, 0),
		uuid:         resUuid,
		eventsChan:   events,
		stop:         true,
	}
}

func (fs *Worker) Run() {
	fs.stop = false
	go fs.readCommands()
	go fs.generateEvents()
}

func (fs *Worker) Stop() {
	fs.stop = true
}

func (fs *Worker) readCommands() {
	_, err := fs.conn.Write([]byte(FsAuthInvite))
	if err != nil {
		return
	}
	buf := ""
	for !fs.stop {
		lBuf := make([]byte, BufLen)
		_ = fs.conn.SetReadDeadline(time.Now().Add(1000000))
		n, err := fs.conn.Read(lBuf)
		_ = fs.conn.SetReadDeadline(time.Unix(0, 0))
		if err != nil {
			// todo add socket errors logging
			continue
		}
		buf = fmt.Sprintf("%s%s", buf, string(lBuf[:n]))
		commands := strings.Split(buf, "\r\n\r\n")
		for i := 0; i < len(commands)-1; i++ {
			go fs.processCommand(commands[i])
		}
		buf = fmt.Sprintf("%s", commands[len(commands)-1:][0])
	}
}

func (fs *Worker) processCommand(s string) {
	msg := strings.Fields(s)
	args := msg[1:]
	switch cmd := msg[0]; cmd {
	case "auth":
		if args[0] == fs.pass {
			if _, err := fs.conn.Write([]byte(FsAuthAcceptedReply)); err != nil {
				fs.stop = true
			}
		} else {
			if _, err := fs.conn.Write([]byte(FsAuthDeniedReply)); err != nil {
				fs.stop = true
			}
			if _, err := fs.conn.Write([]byte(FsDisconnectNotice)); err != nil {
				fs.stop = true
			}
			fs.stop = true
		}
	case "exit":
		if _, err := fs.conn.Write([]byte(FsExitReply)); err != nil {
			fs.stop = true
		}
		if _, err := fs.conn.Write([]byte(FsDisconnectNotice)); err != nil {
			fs.stop = true
		}
		fs.stop = true
	case "event":
		switch args[0] {
		case "plain":
			fs.serialize = SerializePlain
		case "json":
			fs.serialize = SerializeJson
		default:
			if _, err := fs.conn.Write([]byte(FsErrCommandNotFound)); err != nil {
				fs.stop = true
			}
			return
		}
		events := args[1:]
		doCustomEvents := false
		fs.evListsMutex.Lock()
		defer fs.evListsMutex.Unlock()
		for i := range events {
			if len(events[i]) == 0 {
				continue
			}
			if events[i] == "CUSTOM" {
				doCustomEvents = true
				continue
			}
			if !doCustomEvents {
				fs.events = append(fs.events, events[i])
			} else {
				fs.customEvents = append(fs.customEvents, events[i])
			}
		}
	default:
		if _, err := fs.conn.Write([]byte(FsErrCommandNotFound)); err != nil {
			fs.stop = true
		}
	}
}

func (fs *Worker) generateEvents() {
	for !fs.stop {
		select {
		case extEvent := <-fs.eventsChan:
			eventType, err := extEvent.GetHeader("Event-Subtype")
			var list []string
			if err == nil {
				list = fs.customEvents
				fs.evListsMutex.Lock()
			} else {
				list = fs.events
				eventType, _ = extEvent.GetHeader("Event-Name")

			}
			fs.evListsMutex.Lock()
			for i := range list {
				if list[i] == eventType {
					go func() {
						var err error
						switch fs.serialize {
						case SerializePlain:
							err = fs.sendMessage(extEvent.Serialize())
						case SerializeJson:
							err = fs.sendMessage(extEvent.SerializeJson())
						}
						if err != nil {
							fs.stop = true
						}
					}()
				}
			}
			fs.evListsMutex.Unlock()
		default:
			continue
		}
	}
}

func (fs *Worker) sendMessage(buf string) error {
	var tpl string
	switch fs.serialize {
	case SerializePlain:
		tpl = FsPlainEventMessageHeaderTemplate
	case SerializeJson:
		tpl = FsJsonEventMessageHeaderTemplate
	}
	s := fmt.Sprintf(tpl, len(buf)+1)
	if _, err := fs.conn.Write([]byte(s)); err != nil {
		return err
	}
	_, err := fs.conn.Write([]byte(fmt.Sprintf("%s\n", buf)))
	if err != nil {
		return err
	}
	return nil
}
