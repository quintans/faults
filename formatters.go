package faults

import (
	"bytes"
	"fmt"
	"runtime"
)

var formatter Formatter = TextFormatter{}

func SetFormatter(f Formatter) {
	formatter = f
}

type Formatter interface {
	Format(m Message) string
}

type TextFormatter struct{}

func (TextFormatter) Format(m Message) string {
	if !m.Expand {
		return m.Err.Error()
	}

	var trace bytes.Buffer
	trace.WriteString(m.Err.Error())
	for _, frame := range m.Frames() {
		trace.WriteString(fmt.Sprintf("\n    %s:%d", frame.File, frame.Line))
	}
	return trace.String()
}

type Message struct {
	Err    error
	Expand bool
	frames []runtime.Frame
}

func (m Message) Frames() []runtime.Frame {
	return m.frames
}
