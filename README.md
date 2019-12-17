# Just a toolkit for error handling
[![Build Status](https://travis-ci.org/zyguan/just.svg)](https://travis-ci.org/zyguan/just)

Getting tired of keep writing following boilerplate code?

```go
a, err := f(...)
if err != nil {
    return nil, err
}
b, err := g(...)
if err != nil {
    return nil, err
}
```

Just try:

```go
defer just.Return(&err)
a := just.Try(f(...)).Nth(0).(A)
b := just.Try(g(...)).Nth(0).(B)
```

See [print-json-files](examples/print-json-files/main.go) for a complete example.
