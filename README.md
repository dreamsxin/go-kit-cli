This is a fork of the original microgen repo: https://github.com/devimteam/microgen
# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/), by specified service interface.  
The goal is to generate code for service which not fun to write but it should be written.

## Changes from the original repo
1. made the package a go module (added go.mod)
2. removed all dependencies of the $GOPATH - now you can run the tool from any directory
3. added a frienfly prompt in order to help fill the required fields needed to run the tool
4. added defult implementation for converter functions
5. added validation for the interface given - validate the corresponding pb.go file has the same property names and types
6. embeded "unimplementedServer" object to grpc server struct
7. added packgename input
8. added a "message" field to the generated logging middleware
9. added support for stream proto apis
10. reordered the README to show the most important things at the top

## Install
```
go install github.com/dreamsxin/go-kitcli/cmd/microgen@latest
```

Note: If you have problems with building microgen, please clone the repository, and run 
```
go mod vendor
go mod tidy
```

## Usage
``` sh
microgen [OPTIONS]
cd examples/addsvr
go run ..\..\cmd\microgen\main.go -v 100 --file ./addsvc/api.go -out . -main -package github.com/dreamsxin/go-kitcli/examples/addsvc -.proto pb
```
microgen tool search in file first `type * interface` with docs, that contains `// @microgen`.

generation parameters provides through ["tags"](#tags) in interface docs after general `// @microgen` tag (space before @ __required__).

if you run microgen without any flags, the shell will prompt you with insertable fields to insert the required parameters.
microgen will try and guess your output directory and your package name by lokking in the inputed file directory.
example:
``` sh
microgen
@microgen 1.0.0
file path with interfaces: src/service.go
output directory [src]:
pacakge name for imports [github.com/recolabs/reco/auth-service]:
path to XXX_service.pb.go:
all files successfully generated

```

## Interface declaration rules
For correct generation, please, follow rules below.

General:
* Interface should be valid golang code.
* All interface method's arguments and results should be named and should be different (name duplicating unacceptable).
* First argument of each method should be of type `context.Context` (from [standard library](https://golang.org/pkg/context/)).
* Last result should be builtin `error` type.
---
GRPC and Protobuf:  
* Name of _protobuf_ service should be the same, as interface name.
* Function names in _protobuf_ should be the same, as in interface.
* Message names in _protobuf_ should be named `<FunctionName>Request` or `<FunctionName>Response` for request/response message respectively.
* Field names in _protobuf_ messages should be the same, as in interface methods (_protobuf_ - snake_case, interface - camelCase).
---
HTTP GET method (`// @http-method GET`)
* Parameters types should be `string`, `int`, `int32`, `int64`, `uint`, `uint32` or `uint64`.

#### Recommended project layout
Microgen uses [standard-like](https://github.com/golang-standards/project-layout) layout for generating boilerplate.
Default layout of project:
```
├── cmd
│   └── user_service
│       └── main.go
├── pb
│   └── api.proto
├── service
│   ├── caching.go
│   ├── caching.microgen.go
│   ├── error_logging.microgen.go
│   ├── logging.microgen.go
│   ├── middleware.microgen.go
│   └── recovering.microgen.go
├── transport                  // And may be some others in future, NATS or AMQP for example
│   ├── grpc
│   │   ├── client.microgen.go
│   │   ├── protobuf_endpoint_converters.microgen.go
│   │   ├── protobuf_type_converters.microgen.go
│   │   └── server.microgen.go
│   ├── http
│   │   ├── client.microgen.go
│   │   ├── converters.microgen.go
│   │   └── server.microgen.go
│   ├── client.microgen.go
│   ├── endpoints.microgen.go
│   ├── exchanges.microgen.go
│   └── server.microgen.go
├── usersvc
│   ├── api.go
│   └── user.go
├── vendor/
├── Dockerfile
├── Gopkg.lock
├── Gopkg.toml
├── Makefile
└── README.md
```
##### root directory
Contains other dirs and top-level project files, like `Dockerfile, Makefile, Readme.md`, etc.
##### cmd
Contains applications.
##### transport
Contains all transport specific code for all transports: http, grpc, amqp, udp, and so on.
##### service
Middleware in past. Should contain service realisations and closures (middlewares).<br/>
If you need new implementation of service, just add directory `/v2/` or something else.
##### \<project name\>
Contains domain-specific types.
##### and others
Contains everything you want, just separate them by purposes.

Q: How microgen generates for this layout?<br/>
A: `cd <project name>; microgen -out=./..`<br/>

To find more, check examples folder.

### Options

| Name     | Default    | Description                                                                         |
|:---------|:-----------|:------------------------------------------------------------------------------------|
| -file*   |            | Relative path to source file with service interface                                 |
| -out*    |            | Relative or absolute path to directory, where you want to see generated files       |
| -package*|            | Package name for imports                                                            |
| -v       | 1          | Sets microgen verbose level. 0 - print only errors.                                 |
| -help    | false      | Print usage information                                                             |
| -debug   | false      | Print all microgen messages. Equivalent to -v=100.                                  |
| -.proto  |            | Package field in protobuf file. If not empty, service.proto file will be generated. |

\* __Required option__

### Markers
Markers is a general tags, that participate in generation process.
Typical syntax is: `// @<tag-name>:`

#### @microgen
Main tag for microgen tool. Microgen scan file for the first interface which docs contains this tag.  
To add templates for generation, add their [tags](#tags), separated by comma after `@microgen:`
Example:
```go
// @microgen middleware, logging
type StringService interface {
    ServiceMethod(ctx context.Context) (err error)
}
```
#### @protobuf
Protobuf tag is used for package declaration of compiled with `protoc` grpc package.  
Example:
```go
// @microgen grpc-server
// @protobuf github.com/user/repo/path/to/protobuf
type StringService interface {
    ServiceMethod(ctx context.Context) (err error)
}
```
`@protobuf` tag is optional, but required for `grpc`, `grpc-server`, `grpc-client` generation.

#### @grpc-addr
This tag allows to add construction for default grpc server addr in generated grpc client.
```go
if addr == "" {
    addr = "service.string.StringService"
}
```
Example:
```go
// @microgen grpc-client
// @grpc-addr service.string.StringService
type StringService interface {
    ServiceMethod(ctx context.Context) (err error)
}
```

### Method's tags
#### @microgen one-to-many
Microgen will treat this function as a one to many stream api.

```go
// @microgen main, logging
type StringService interface {
    // @microgen one-to-many
    Count(text string, symbol string, stream v1.StringService_CountServer) (err error)
}
```
#### @microgen many-to-many
Microgen will treat this function as a many to many stream api.

```go
// @microgen main, logging
type StringService interface {
    // @microgen many-to-many
    Count(stream v1.StringService_CountServer) (err error)
}
```

#### @microgen many-to-one
Microgen will treat this function as a many to one stream api.

```go
// @microgen main, logging
type StringService interface {
    // @microgen many-to-one
    Count(stream v1.StringService_CountServer) (err error)
}
```
#### @microgen -
Microgen will ignore method with this tag everywere it can.

```go
// @microgen main, logging
type StringService interface {
    // @microgen -
    Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
}
```

#### @logs-ignore
This tag is used for logging middleware, when some arguments or results should not be logged, e.g. passwords or files.  
If `context.Context` is first argument, it ignored by default.
Provide parameters names, separated by comma, to exclude them from logs.  
Example:
```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

#### @logs-len
This tag is used for logging middleware. It prints length of parameters.
Example:  
```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    // @logs-len data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

#### @http-method
This tag is used for http server and client to set different from POST method. This tag has special validation rules for GET method.
Example:  
```go
// @microgen logging
type StringService interface {
    // @http-method GET
    Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
}
```

#### cache-key
This tag is used for caching middleware and allows user to write expression that should be used as key for cache instance.<br/>
Key may be any string: it will directly writes to generated code.

```go
// @microgen caching
type StringService interface {
    // @cache-key strings.ToLower(text)
    Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
}
```

### Tags
All allowed tags for customize generation provided here.

| Tag         | Description                                                                                                                   |
|:------------|:------------------------------------------------------------------------------------------------------------------------------|
| middleware  | General application middleware interface. Generates every time.                                                               |
| logging     | Middleware that writes to logger all request/response information with handled time. Generates every time.                    |
| error-logging | Middleware that writes to logger errors of method calls, if error is not nil.                                               |
| recovering  | Middleware that recovers panics and writes errors to logger. Generates every time.                                            |
| caching     | Middleware that caches responses of service. Adds missed functions.                                                           |
| grpc-client | Generates client for grpc transport with request/response encoders/decoders. Do not generates again if file exist.            |
| grpc-server | Generates server for grpc transport with request/response encoders/decoders. Do not generates again if file exist.            |
| grpc        | Generates client and server for grpc transport with request/response encoders/decoders. Do not generates again if file exist. |
| http-client | Generates client for http transport with request/response encoders/decoders. Do not generates again if file exist.            |
| http-server | Generates server for http transport with request/response encoders/decoders. Do not generates again if file exist.            |
| http        | Generates client and server for http transport with request/response encoders/decoders. Do not generates again if file exist. |
| main        | Generates basic `package main` for starting service. Uses other tags for minimal user changes.                                |
| tracing     | Generates options and params for opentracing.                                                                                 |
| metrics     | Generates transport endpoints middlewares for common tracing purposes.                                                                                 |

## Example
You may find examples in `examples` directory, where `svc` contains all, what you need for successful generation, and `generated` contains what you will get after `microgen`.

Follow this short guide to try microgen tool.

1. Create file `service.go` inside GOPATH and add code below.
```go
package stringsvc

import (
	"context"

	"github.com/dreamsxin/go-kitcli/example/svc/entity"
)

// @microgen middleware, logging, grpc, http, recovering, main
// @protobuf github.com/dreamsxin/go-kitcli/example/protobuf
type StringService interface {
	// @logs-ignore ans, err
	Uppercase(ctx context.Context, stringsMap map[string]string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
	// @logs-len comments
	TestCase(ctx context.Context, comments []*entity.Comment) (tree map[string]int, err error)
}
```
2. Open command line next to your `service.go`.
3. Enter `microgen`. __*__
4. You should see something like that:
```
@microgen 0.5.0
all files successfully generated
```
5. Now, add and generate protobuf file (if you use grpc transport) and write transport converters (from protobuf/json to golang and _vise versa_).
6. Use endpoints in your `package main` or wherever you want. (tag `main` generates some code for `package main`)

__*__ `GOPATH/bin` should be in your PATH.

## Dependency
list out of date!
After generation your service may depend on this packages:
```
    "net/http"      // for http purposes
    "bytes"
    "encoding/json" // for http purposes
    "io/ioutil"
    "strings"
    "net"           // for http and grpc listners
    "net/url"       // for http purposes
    "fmt"
    "context"
    "time"          // for logging
    "os"            // for signal handling and os.Stdout
    "os/signal"     // for signal handling 
    "syscall"       // for signal handling
    "errors"        // for error handling
    

    "google.golang.org/grpc"                    // for grpc purposes
    "golang.org/x/net/context"
    "github.com/go-kit/kit"                     // for grpc purposes
    empty "google.golang.org/protobuf/types/known/emptypb"   // for grpc purposes
```
