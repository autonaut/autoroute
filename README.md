autoroute
------

build wicked fast, automatically documented APIs with Go.

Autoroute uses reflection to automatically create JSON APIs from certain classes of Go functions.

Mix and match any of the following input parameters ->

```go
func(ctx context.Context, header autoroute.Header, input anyJSONDecodeableStructOrStructPointer)
func(ctx context.Context, input anyJSONDecodeableStructOrStructPointer)
func(ctx context.Context, header autoroute.Header)
func(input anyJSONDecodeableStructOrStructPointer)
func(ctx context.Context)
```

with any of the following output paramters.

```go
error
anyJSONEncodableStructOrStructPointer
(error, anyJSONEncodableStructOrStructPointer)
void
```

```go
type DemoInputStruct {
    Email string `json:"email" form:"email" valid:"isEmail"`
}
```


LICENSE
======

MIT, see LICENSE