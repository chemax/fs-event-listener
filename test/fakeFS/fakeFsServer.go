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
package event_listener_test

import (
	uuid2 "github.com/google/uuid"
	"net"
)

type worker struct {
	fs         Worker
	eventsChan chan *Event
}

type Server struct {
	addr       string
	password   string
	uuid       string
	workers    []worker
	listener   net.Listener
	eventsList []*Event
	stop       bool
}

func NewServer(addr string, password string, events []*Event) (*Server, string, error) {
	var (
		err      error
		listener net.Listener
	)
	uuid, _ := uuid2.NewUUID()
	if listener, err = net.Listen("tcp", addr); err != nil {
		return nil, uuid.String(), err
	}
	server := &Server{
		addr:       addr,
		password:   password,
		uuid:       uuid.String(),
		listener:   listener,
		stop:       false,
		eventsList: events,
	}
	go server.startServeConnections()
	go server.startEventGenerator()
	return server, uuid.String(), nil
}

func (s *Server) Stop() {
	if err := s.listener.Close(); err != nil {
		// todo add error loggigng
	}
	s.stop = true
}

func (s *Server) startServeConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			break
		}
		eventsChan := make(chan *Event)
		servInstance := *NewWorker(conn, s.password, s.uuid, eventsChan)
		lworker := worker{
			fs:         servInstance,
			eventsChan: eventsChan,
		}
		s.workers = append(s.workers, lworker)
		lworker.fs.Run()
	}
}

func (s *Server) startEventGenerator() {
	for !s.stop {
		for i := range s.eventsList {
			for k := range s.workers {
				s.workers[k].eventsChan <- s.eventsList[i]
			}
		}
	}
}
