package fakeFS

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
