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
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Event struct {
	body    string
	headers []string
}

func NewEvent(eventName string) *Event {
	baseHeader := make([]string, 0)
	if strings.Index(eventName, "CUSTOM") > -1 {
		baseHeader = append(baseHeader, "Event-Name: CUSTOM")
		baseHeader = append(baseHeader, fmt.Sprintf("Event-Subclass: %s", strings.Fields(eventName)[1]))
	} else {
		baseHeader = append(baseHeader, fmt.Sprintf("Event-Name: %s", eventName))
	}
	return &Event{
		headers: baseHeader,
		body:    "",
	}
}

func (e *Event) Serialize() string {
	result := ""
	for i := range e.headers {
		result = fmt.Sprintf("%s%s\n", result, e.headers[i])
	}
	if len(e.body) > 0 {
		result = fmt.Sprintf("%sContent-Length: %v\n\n%s", result, len(e.body), e.body)
	}
	return result
}

func (e *Event) SerializeJson() string {
	var result []byte
	res := make(map[string]string)
	for i := range e.headers {
		s := strings.SplitN(e.headers[i], ":", 2)
		// fixme
		if s[1][0] == ' ' {
			res[s[0]] = s[1][1:]
		} else {
			res[s[0]] = s[1]
		}
	}
	if len(e.body) > 0 {
		res["body"] = e.body
	}
	var err error
	result, err = json.Marshal(res)
	if err != nil {
		log.Printf("Error while converting event to JSON: %v\n", err)
	}
	return string(result)
}

func (e *Event) SetHeader(name, value string) {
	for i := range e.headers {
		if strings.Index(e.headers[i], fmt.Sprintf("%s:", name)) == 0 {
			e.headers[i] = fmt.Sprintf("%s: %s", name, value)
			return
		}
	}
	e.headers = append(e.headers, fmt.Sprintf("%s: %s", name, value))
}

type GetHeaderError struct {
	hname string
}

func (e GetHeaderError) Error() string {
	return fmt.Sprintf("Error while getting header %s!", e.hname)
}

func (e *Event) GetHeader(name string) (string, error) {
	for i := range e.headers {
		if strings.Index(e.headers[i], fmt.Sprintf("%s:", name)) == 0 {
			return strings.Fields(e.headers[i])[1], nil
		}
	}
	return "", GetHeaderError{hname: name}
}

func (e *Event) AddBody(body string) {
	if len(e.body) > 0 {
		e.body = fmt.Sprintf("%s%s", e.body, body)
	} else {
		e.body = body
	}
}
