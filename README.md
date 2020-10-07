# errors
Error wrapping package that provides extra context.

# Examples
```go
err := strconv.Itoa("z")
if err != nil {
  return errors.E(errors.CodeServerError, errors.Decoding, err)
}
```


```go
accountid := 0
if accountid == 0 {
  return errors.E(errors.CodeInvalid, errors.Invalid, "Invalid Account ID")
}
```

This will return an error that is easily json marshaled and allows for stack tracing like so

```json
{
  "code": 422,
  "op": "",
  "kind": 1,
  "err": "Invalid Account ID",
  "stack": [
    "test/main.go:38 return nil, errors.E(errors.CodeInvalid, errors.Invalid, \"Invalid Account ID\")",
    "github.com/hunterdishner/gomux@v0.0.0-20201007225513-ee88fc69884c/gomux.go:231 data, err := fn(w, r)",
    "net/http/server.go:2012 f(w, r)"
  ]
}
```
