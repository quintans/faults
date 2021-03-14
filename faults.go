package faults

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

const (
	maxStackLength = 50
	callerOffset   = 4
)

var formatter Formatter = TextFormatter{}

func SetFormatter(f Formatter) {
	formatter = f
}

type Formatter interface {
	Format(err error, frames []runtime.Frame) string
}

type TextFormatter struct{}

func (TextFormatter) Format(err error, frames []runtime.Frame) string {
	var trace bytes.Buffer
	trace.WriteString(fmt.Sprintf("%+v", err))
	for _, frame := range frames {
		trace.WriteString(fmt.Sprintf("\n  %s:%d", frame.File, frame.Line))
	}
	return trace.String()
}

// Error is the type that implements the error interface.
// It contains the underlying err and its stacktrace.
type Error struct {
	Err        error
	stack      []uintptr
	stackTrace string
}

// Unwrap unpacks wrapped errors
func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Format(s fmt.State, verb rune) {
	if verb != 'v' || !s.Flag('+') || e.stack == nil {
		s.Write([]byte(e.Error()))
		return
	}

	if e.stackTrace == "" {
		e.stackTrace = getStackTrace(e.Err, e.stack)
	}
	s.Write([]byte(e.stackTrace))
}

func getStackTrace(err error, stack []uintptr) string {
	frames := runtime.CallersFrames(stack)
	var fs []runtime.Frame
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			fs = append(fs, frame)
		}
		if !more {
			break
		}
	}
	return formatter.Format(err, fs)
}

// New returns a new error creates a new
func New(text string) error {
	return wrap(errors.New(text), 0)
}

// Errorf creates a new error based on format and wraps it in a stack trace.
// The format string can include the %w verb.
func Errorf(format string, args ...interface{}) error {
	return wrap(fmt.Errorf(format, args...), 0)
}

// Wrap annotates the given error with a stack trace
func Wrap(err error) error {
	return wrap(err, 0)
}

// WrapUp to be used by custom utility functions
func WrapUp(err error) error {
	return wrap(err, 1)
}

func wrap(err error, offset int) error {
	if err == nil {
		return err
	}

	switch err.(type) {
	case *Error:
		return err
	}

	var e *Error
	if errors.As(err, &e) {
		// keeping the stack in the top level error
		newErr := &Error{
			Err:   err,
			stack: e.stack,
		}
		// reset
		*e = Error{Err: e.Err}
		return newErr
	}

	return &Error{Err: err, stack: getStack(offset)}
}

func getStack(offset int) []uintptr {
	stackBuf := make([]uintptr, maxStackLength)
	length := runtime.Callers(callerOffset+offset, stackBuf[:])
	return stackBuf[:length]
}

func Trace(errp *error, format string, args ...interface{}) {
	if *errp == nil {
		return
	}

	s := fmt.Sprintf(format, args...)
	*errp = wrap(fmt.Errorf("%s: %w", s, *errp), 0)
}
