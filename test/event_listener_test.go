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
	"github.com/0x19/goesl"
	EL "github.com/borikinternet/fs-event-listener"
	FS "github.com/borikinternet/fs-event-listener/test/fakeFS"
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
	time.Sleep(time.Millisecond * 100)
	if !success {
		t.Fail()
	}
}

func TestCustomEventListener(t *testing.T) {
	runtime.GOMAXPROCS(2)
	eventTest := FS.NewEvent("CUSTOM test::test")
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
	eListener.AddEventHandler("CUSTOM test::test", func(event *goesl.Message) {
		if event.GetHeader("Event-Name") == "CUSTOM" && event.GetHeader("Event-Subclass") == "test::test" {
			success = true
		}
	})
	if err := eListener.OpenESLConnection("127.0.0.1", "ClueCon", 8021, 1); err != nil {
		t.Fail()
		return
	}
	time.Sleep(time.Millisecond * 100)
	if !success {
		t.Fail()
	}
}
