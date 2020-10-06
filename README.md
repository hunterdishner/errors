# errors
Error wrapping package that provides extra context.

# Examples
```
err := strconv.Itoa("z")
if err != nil {
  return errors.E(errors.CodeServerError, errors.Decoding, err)
}
```


```
accountid := 0
if accountid == 0 {
  return errors.E(errors.CodeInvalid, errors.Invalid, "Invalid Account ID")
}
```

This will return an error that is easily json marshaled and allows for stack tracing. 
