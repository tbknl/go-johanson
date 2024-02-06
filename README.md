go-johanson
===========

JSON stream writing Go module supporting a convenient workflow for data without corresponding (tagged) structs.

```shell
go get github.com/tbknl/go-johanson
```

## Rationale

This module helps out in situations where JSON needs to be written to a stream, while
* no (tagged) structs are present in the code, which can be used to marshal the data;
* not all data is available in a single structure to serialize in one go;
* the amount of data is to large to buffer completely in memory beforing starting to write;
* you prefer a programmatic way to define the JSON data structure.

## Simple example
```go
import "github.com/tbknl/go-johanson"

func main() {
    // NOTE: `writer` can be any object implementing the `io.Writer` interface,
    //       for example `os.Stdout`, `&strings.Builder` and `http.ResponseWriter`.
    writer := os.Stdout
    jsw := johanson.NewStreamWriter(writer)

    jsw.Array(func(a johanson.V) {
        a.Uint(123)
        a.String("Hello")
        a.Object(func(o johanson.K) {
            o.Item("str").String("value1")
            o.Item("float").Float(45.67)
            o.Item("null").Null()
        })
        a.Bool(true)
        a.Int(-999)
        a.Marshal(map[string]interface{}{ "one": 1, "two": []int{2} })
    })
}
```

The above code will write the following to stdout: `[123,"Hello",{"str":"value1","float":45.67,"null":null},true,-999,{"one":1,"two":[2]}]`

## Example with an http handler
```go
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "application/json")
    jsw := johanson.NewStreamWriter(w)
    jsw.Object(func(o johanson.K) {
        o.Item("data").String("My json data!")
    })
}
```

## Features

* Writes JSON to any stream directly.
* String escaping and float formatting using the built-in `encoding/json` module (for now).
* Support for marshaling, which uses the `encoding/json` module and therefore can encode anything which this module can.

## API reference

### Create a new JSON stream writer

```go
writer := os.Stdout
jsw := johanson.NewStreamWriter(writer)
```

### Basic data types

* `.Null()` writes `null` to the stream.
* `.Bool(value bool)` writes `true` or `false` to the stream.
* `.Int(value int64)` writes the integer value to the stream.
* `.Uint(value uint64)` writes the unsigned integer value to the stream.
* `.Float(value float64)` writes the floating point value to the stream.
* `.String(value string)` writes the string value to the stream, escaping appropriate characters.

### Arrays

* `.Array(func(arr johanson.V) { /* ... */ })` opens an array context and allows writing elements to it within the callback function.

### Objects

* `.Object(func(obj johanson.K) { /* ... */ })` opens an object context and allows writing items to it within the callback function.
    * `obj.Item(key string) johanson.V` opens an object-item context, allowing to write a value to the object with the given key.
    * `obj.Marsha(object map[string]interface{}) error` marshals the provided object/map, and writes its items (if any) to the stream, returning an error if something went wrong with the marshaling.

### Marshaling

* `.Marshal(value interface{}) error` invokes the `encoding/json` function `json.Marshal(value)` to write arbitrary data to the stream, returning an error if something went wrong (typically only on non-marshalable data).

### Supporting functions

* `.Finished() bool` checks whether the stream writer is finished.
* `.Error() error` returns the last error that occurred when writing to the stream, or `nil` if no error occurred.

