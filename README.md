# faults
Wrap errors with stack trace

## Usage

### Errorf

The most common usage should be `faults.Errorf` that can be a replacement for `fmt.Errorf`

```go
err := errors.New("Bad data") // created by a call to an external lib
err = faults.Errorf("Unable to process data: %w", err)
```

The output of `err.Error()` will be the same as `fmt.Errorf`.

```sh
Unable to process data: Bad data
```

To print the stack trace we have to use `%+v`.

```go
fmt.Printf("%+v\n", err)
```

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

### Trace

utility function to be used in our function calls to trace call values, for example.

```go
func doAnotherStuff(b int) (err error) {
	defer Trace(&err, "doAnotherStuff(b=%d)", b)

	if b <= 0 {
		return ErrInvalidArgument
	}

	return nil
}
```

on error this would output

```sh
doAnotherStuff(b=-1): invalid argument
	/.../app/stuff.go:28
	/.../app/caller.go:123
```

### WrapUp

starts the caller stack trace one level up.
This allows us to remove the extra stack trace line if we want to use this lib in custom utility code.

eg: write a different `Trace()` method.

