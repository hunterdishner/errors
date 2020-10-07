# Errors
Error wrapping package that provides more verbose errors and allows for stack traces

## Examples

Most all errors returned will be wrapped in the errors.E function which will give you back an errors.Error that satisfies the requirements of the standard library error interface.

```go
err := strconv.Itoa("z")
if err != nil {
  return errors.E(errors.CodeServerError, errors.Decoding, err)
}
```
The returned value, when json marshaled, will look like so 

```json
{
  "code": 500,
  "op": "",
  "kind": 9,
  "err": "strconv.Atoi: parsing \"z\": invalid syntax",
  "stack": [
    "test/main.go:39 return nil, errors.E(errors.CodeServerError, errors.Decoding, err)",
    "github.com/hunterdishner/gomux@v0.0.0-20201007225513-ee88fc69884c/gomux.go:231 data, err := fn(w, r)",
    "net/http/server.go:2012 f(w, r)"
  ]
}
```

Notice how the original error message is preserved in the new error struct

---

```go
accountid := 0
if accountid == 0 {
  return errors.E(errors.CodeInvalid, errors.Invalid, "Invalid Account ID")
}
```

This snippet will return an error that is marshaled to this

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

Note: The error message value is the string that was passed in at the end.
