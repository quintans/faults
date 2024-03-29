package faults

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

const (
	maxStackLength = 50
	callerOffset   = 4
)

// rootError is the type that implements the error interface.
// It contains the underlying err and its stacktrace.
type rootError struct {
	err    error
	stack  []uintptr
	frames []runtime.Frame
}

func (m *rootError) Frames() []runtime.Frame {
	if m.frames != nil {
		return m.frames
	}

	if m.stack == nil {
		return nil
	}

	frames := runtime.CallersFrames(m.stack)
	var fs []runtime.Frame
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			break
		}
		fs = append(fs, frame)
		if !more {
			break
		}
	}

	m.frames = fs
	return fs
}

// Unwrap unpacks wrapped errors
func (e *rootError) Unwrap() error {
	return e.err
}

func (e *rootError) Error() string {
	return e.err.Error()
}

type wrapError struct {
	err error
}

// Unwrap unpacks wrapped errors
func (e *wrapError) Unwrap() error {
	return e.err
}

func (e *wrapError) Error() string {
	return e.err.Error()
}

func (e *wrapError) Format(s fmt.State, verb rune) {
	var r *rootError
	var text string

	if errors.As(e.err, &r) {
		expand := verb == 'v' && s.Flag('+') && r.stack != nil
		text = formatter.Format(Message{Expand: expand, Err: e.err, frames: r.Frames()})
	} else {
		text = e.err.Error()
	}

	s.Write([]byte(text))
}

func ToString(err error) string {
	var r *rootError

	if errors.As(err, &r) {
		expand := r.stack != nil
		return formatter.Format(Message{Expand: expand, Err: err, frames: r.Frames()})
	}

	return err.Error()
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

// Wrapf creates a new error based on format and wraps it in a stack trace,
// only if error is not nil
// The format string cannot include the %w verb.
// error will concatenated to format like: format + ": %w"
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	args = append(args, err)
	return wrap(fmt.Errorf(format+": %w", args...), 0)
}

// WrapUp to be used by custom utility functions
func WrapUp(err error) error {
	return wrap(err, 1)
}

func wrap(err error, offset int) error {
	if err == nil {
		return nil
	}

	var e *wrapError
	if errors.As(err, &e) {
		return &wrapError{err}
	}

	return &wrapError{&rootError{err: err, stack: getStack(offset)}}
}

func getStack(offset int) []uintptr {
	stackBuf := make([]uintptr, maxStackLength)
	length := runtime.Callers(callerOffset+offset, stackBuf)
	if length == 0 {
		return nil
	}
	return stackBuf[:length]
}

// Catch allow us to implement a different pattern of error tracing where it is the responsibility of the callee to say that it was called.
// This allows us to not worry about adding an message on every error return.
//
//	func (o *Something) MethodM(a TypeA) (_ Stuff, er error) {
//		defer faults.Catch(&er, "calling MethodM(a=%v)", a)
//
//		x, err := doSomething(a) // will also have a Catch() inside
//		if err != nil {
//			return Stuff{}, faults.Wrap(err)
//		}
//		s, err := stuff.New(a, x)  // will also have a Catch() inside
//		if err != nil {
//			return Stuff{}, faults.Wrap(err)
//		}
//		return s, nil
//	}
//
// where the error message would be: "calling MethodM(a=-2): calling doSomething(a=-2): value must > 0"
func Catch(errp *error, format string, args ...interface{}) {
	if *errp == nil {
		return
	}

	s := fmt.Sprintf(format, args...)
	*errp = wrap(fmt.Errorf("%s: %w", s, *errp), 0)
}
