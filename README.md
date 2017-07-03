# Just a toolkit for error handling
[![Build Status](https://travis-ci.org/zyguan/just.svg)](https://travis-ci.org/zyguan/just)

Getting tired of keep writing following boilerplate code?

```go
ret, err := f(...)
if err != nil {
    return nil, err
}
```

Just try:

```go
defer just.Catch(&err)
ret := just.Try(f(...)).(A)
```

See [openall](examples/openall/main.go) for a slightly complex example.