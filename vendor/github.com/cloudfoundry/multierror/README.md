# Multierror

`multierror` is a simple go package that allows you to combine and present multiple errors as a single error.

# Installation
Run `go get github.com/cloudfoundry/multierror`

# How to use

```go
import "github.com/cloudfoundry/multierror"

errors := multierror.MultiError{}

err1 := FirstFuncThatReturnsError()
err2 := SecondFuncThatReturnsError()

errors.Add(err1)
errors.Add(err2)

//You can also add multierror structs and they will be flattened into one struct
errors2 := multierror.MultiError()
errors2.Add(err1)
errors.Add(errors2)


//Returns the errors as an aggregate of all the error messages
errors.Error()
```


