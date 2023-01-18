package faults

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		format string
		msg    string
	}{
		{
			name: "plain error",
			err:  errors.New("plain"),
			msg:  "plain",
		},
		{
			name:   "composit plain error",
			err:    errors.New("something"),
			format: "This has a message: %w",
			msg:    "This has a message: something",
		},
		{
			name:   "lib plain error",
			err:    New("something"),
			format: "This has a message: %w",
			msg:    "This has a message: something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.format == "" {
				err = Wrap(tt.err)
			} else {
				err = Errorf(tt.format, tt.err)
			}

			err = fmt.Errorf("double wrapping: %w", err)
			err = Wrap(err)

			expect := "double wrapping: " + tt.msg
			assert.Equal(t, expect, fmt.Sprintf("%s", err))
			assert.Equal(t, expect, err.Error())
			full := fmt.Sprintf("%+v", err)
			assert.True(t, strings.HasPrefix(full, expect), full)
			assert.Equal(t, 3, countLines(full), full)
		})
	}
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func TestTrace(t *testing.T) {
	err := doStuff("Hello", 1)
	require.NoError(t, err)

	err = doStuff("World", -1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "doStuff(a=World, b=-1): doAnotherStuff(b=-1): invalid argument")
}

var ErrInvalidArgument = errors.New("invalid argument")

func doStuff(a string, b int) (err error) {
	defer Catch(&err, "doStuff(a=%s, b=%d)", a, b)

	return doAnotherStuff(b)
}

func doAnotherStuff(b int) (err error) {
	defer Catch(&err, "doAnotherStuff(b=%d)", b)

	if b <= 0 {
		return ErrInvalidArgument
	}

	return nil
}
