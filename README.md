# faults
Wrap errors with stack trace

## Usage

### Errorf

The most common usage should be `faults.Errorf` that can be a replacement for `fmt.Errorf`

```go
err := errors.New("Bad data") // created by a call to an external lib
err = faults.Errorf("Unable to process data: %w", err)
```

The output of `err.Error()` will be the same as `fmt.Errorf` but with an added stack trace.

```sh
Unable to process data: Bad data
	/.../app/process.go:28
	/.../app/caller.go:1123
```

Additional calls to `faults.Errorf` will not change the stack trace.

### Wrap

We can also wrap any error with `faults.Wrap(err)`

```go
err := errors.New("Bad data") // created by a call to an external lib
err = faults.Wrap(err)
```

If the error is nil or is already wrapped, no action will be taken, otherwise the error will be wrapped with a stack trace info.
This means that we can just use `faults.Wrap(err)` on any error without thinking if it is already wrapped or not.

### New

If we have to create a new error on the fly, we can use `faults.New`

```go
return faults.New("Bad data")
```

> Don't use `faults.New` when declaring a global variable because the stack trace will be relatively to the point of declaration
