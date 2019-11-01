package event_listener_test

import (
	EL "../eventListener"
	FS "./fakeFS"
	"github.com/0x19/goesl"
	"runtime"
	"testing"
	"time"
)

func TestEventListener(t *testing.T) {
	runtime.GOMAXPROCS(2)
	eventTest := FS.NewEvent("TEST")
	eventsList := make([]*FS.Event, 0)
	eventsList = append(eventsList, eventTest)
	fs, uuid, err := FS.NewServer("127.0.0.1:8021", "ClueCon", eventsList)
	if err != nil {
		// todo add error handling
		t.Fail()
		return
	}
	defer fs.Stop()
	eventTest.SetHeader("Core-UUID", uuid)
	eListener := EL.NewEventListener()
	success := false
	eListener.AddEventHandler("TEST", func(event *goesl.Message) {
		if event.GetHeader("Event-Name") == "TEST" {
			success = true
		}
	})
	if err := eListener.OpenESLConnection("127.0.0.1", "ClueCon", 8021, 1); err != nil {
		t.Fail()
		return
	}
	time.Sleep(100000000)
	if !success {
		t.Fail()
	}
}
