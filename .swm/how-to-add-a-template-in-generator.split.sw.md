---
id: split
name: How to Add a Template in Generator
file_version: 1.0.2
app_version: 0.9.8-3
file_blobs:
  generator/template/service_logging.go: 04833d590d16847e95522302bc6df1eb2c35dd45
  generator/template/template.go: 4c2ccb776212dae4cf1164d87a8d296accdccf02
  generator/write_strategy/file.go: c49c0553a078283171ee198dbe7118a6266b4eb3
---

In this document, we will learn how to add a new Template to the system. Templates, like [Golang's text/template engine](https://pkg.go.dev/text/template), are ways to generate textual output in a data-driven way.

You'd want to add a template if you find another part of your μServices often repeats and can be auto-generated.

Some examples of `📄 generator/template`s are `📄 generator/template/buffer_adapter.go`, `📄 generator/template/cmd_main.go`, `📄 generator/template/service_middleware.go`, and `📄 generator/template/common.go`.

## TL;DR - How to add `📄 generator/template`s

1.  Create a new file under `📄 generator/template` 
    
    *   e.g. `📄 generator/template/service_logging.go`
        
2.  Add a struct type in your new file, i.e.:
    
    ```golang
    type XXXTemplate struct {
    // The struct should hold the data you need to render your template
    }
    ```
    
    

3.  You'll need to implement the interface `Template`[<sup id="1ghvhI">↓</sup>](#f-1ghvhI), which means you need to implement 4 methods on your struct type: `Prepare`[<sup id="2n3btL">↓</sup>](#f-2n3btL), `DefaultPath`[<sup id="ZRpVyT">↓</sup>](#f-ZRpVyT), `ChooseStrategy`[<sup id="Z1tc5zk">↓</sup>](#f-Z1tc5zk), and most importantly: `Render`[<sup id="Z2dN3kR">↓</sup>](#f-Z2dN3kR).
    
4.  **Profit** 💰
    

# Full walkthrough

Start by creating a new file under `📄 generator/template`. We'll follow `📄 generator/template/service_logging.go` as an example.

## File boilerplate

Like any Go file, you need to specify the package and the imports. This is not super interesting and your IDE can probably do it automatically for you, but here is how we do it for `📄 generator/template/service_logging.go`:

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/service_logging.go
```go
🟩 1      package template
🟩 2      
🟩 3      import (
🟩 4      	"context"
🟩 5      
🟩 6      	. "github.com/dave/jennifer/jen"
🟩 7      	mstrings "github.com/dreamsxin/go-kitcli/generator/strings"
🟩 8      	"github.com/dreamsxin/go-kitcli/generator/write_strategy"
🟩 9      	"github.com/vetcher/go-astra/types"
🟩 10     )
```

<br/>

## Implementing `Prepare`[<sup id="2n3btL">↓</sup>](#f-2n3btL)

Prepare is all about prepping the data (for example, scanning files) so that the Render can happen - rendering is generating data-driven files, after all.

Here's our example of Prepare. In this case, we need to know which parameters we want to ignore, so we won't log them. For this middleware, this is useful to hide parameters that might contain secrets from getting to the logfile and f\*\*king up your SOC2 compliance report when the auditors find passwords in Datadog.

<br/>

<div align="center"><img src="https://media4.giphy.com/media/dU0MSRsHl1zPEP55zG/giphy.gif?cid=d56c4a8bbw3f7e8f11zuup6szl1yhsg60nid5mtlch2q7hg2&rid=giphy.gif&ct=g" style="width:'50%'"/></div>

<br/>

So we're initializing the `ignoreParams`[<sup id="ZROtS5">↓</sup>](#f-ZROtS5) member with an empty map, and populating the map with all the parameters we want to ignore (based on the relevant struct tags).

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/service_logging.go
```go
⬜ 115    	return filenameBuilder(PathService, "logging")
⬜ 116    }
⬜ 117    
🟩 118    func (t *loggingTemplate) Prepare(ctx context.Context) error {
🟩 119    	t.ignoreParams = make(map[string][]string)
🟩 120    	t.lenParams = make(map[string][]string)
🟩 121    	for _, fn := range t.info.Iface.Methods {
🟩 122    		t.ignoreParams[fn.Name] = mstrings.FetchTags(fn.Docs, TagMark+logIgnoreTag)
🟩 123    		t.lenParams[fn.Name] = mstrings.FetchTags(fn.Docs, TagMark+lenTag)
🟩 124    	}
🟩 125    	return nil
🟩 126    }
🟩 127    
⬜ 128    func (t *loggingTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
⬜ 129    	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
⬜ 130    }
```

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/template.go
```go
🟩 9      type Template interface {
🟩 10     	// Do all preparing actions, e.g. scan file.
🟩 11     	// Should be called first.
🟩 12     	Prepare(ctx context.Context) error
🟩 13     	// Default relative path for template (=file)
🟩 14     	DefaultPath() string
🟩 15     	// Template chooses generation strategy, e.g. appends to file or create new.
🟩 16     	ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error)
🟩 17     	// Main render function, where template produce code.
🟩 18     	Render(ctx context.Context) write_strategy.Renderer
🟩 19     }
🟩 20     
```

<br/>

## Implementing `DefaultPath`[<sup id="ZRpVyT">↓</sup>](#f-ZRpVyT) & `ChooseStrategy`[<sup id="Z1tc5zk">↓</sup>](#f-Z1tc5zk)

This part should be a lot easier.

First, implement `DefaultPath`[<sup id="ZRpVyT">↓</sup>](#f-ZRpVyT) so that the renderer will know where to put the generated files at.

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/service_logging.go
```go
⬜ 113    
🟩 114    func (loggingTemplate) DefaultPath() string {
🟩 115    	return filenameBuilder(PathService, "logging")
🟩 116    }
⬜ 117    
```

<br/>

Then, implement `ChooseStrategy`[<sup id="Z1tc5zk">↓</sup>](#f-Z1tc5zk), which informs the renderer whether to create a new file or append to an existing file. A safe bet would be to just use the `NewCreateFileStrategy`[<sup id="Z1eqtoA">↓</sup>](#f-Z1eqtoA).

## Implementing `Render`[<sup id="Z2dN3kR">↓</sup>](#f-Z2dN3kR) - what we're all here for

<br/>

<div align="center"><img src="https://media0.giphy.com/media/xT39Db8zIOODTppk08/giphy.gif?cid=d56c4a8batg154a1mlnihllndm90znoxtfan7z0rc3frbuck&rid=giphy.gif&ct=g" style="width:'50%'"/></div>

<br/>

This is the fun part! First, read the example output. This is how the output of Render _SHOULD_ look like:

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/service_logging.go
```go
⬜ 33     
⬜ 34     // Render all logging.go file.
⬜ 35     //
🟩 36     //		// This file was automatically generated by "microgen" utility.
🟩 37     //		// DO NOT EDIT.
🟩 38     //		package middleware
🟩 39     //
🟩 40     //		import (
🟩 41     //			context "context"
🟩 42     //			svc "github.com/dreamsxin/go-kitcli/examples/svc"
🟩 43     //			log "github.com/go-kit/kit/log"
🟩 44     //			time "time"
🟩 45     //		)
🟩 46     //
🟩 47     //		func ServiceLogging(logger log.Logger) Middleware {
🟩 48     //			return func(next svc.StringService) svc.StringService {
🟩 49     //				return &serviceLogging{
🟩 50     //					logger: logger,
🟩 51     //					next:   next,
🟩 52     //				}
🟩 53     //			}
🟩 54     //		}
🟩 55     //
🟩 56     //		type serviceLogging struct {
🟩 57     //			logger log.Logger
🟩 58     //			next   svc.StringService
🟩 59     //		}
🟩 60     //
🟩 61     //		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
🟩 62     //			defer func(begin time.Time) {
🟩 63     //				s.logger.Log(
🟩 64     //					"method", "Count",
🟩 65     //					"text", text,
🟩 66     // 					"symbol", symbol,
🟩 67     //					"count", count,
🟩 68     // 					"positions", positions,
🟩 69     //					"took", time.Since(begin))
🟩 70     //			}(time.Now())
🟩 71     //			return s.next.Count(ctx, text, symbol)
🟩 72     //		}
⬜ 73     //
⬜ 74     func (t *loggingTemplate) Render(ctx context.Context) write_strategy.Renderer {
⬜ 75     	f := NewFile("service")
```

<br/>

Now, review the implementation that created this output.

<br/>



<!-- NOTE-swimm-snippet: the lines below link your snippet to Swimm -->
### 📄 generator/template/service_logging.go
```go
🟩 74     func (t *loggingTemplate) Render(ctx context.Context) write_strategy.Renderer {
🟩 75     	f := NewFile("service")
🟩 76     	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
🟩 77     	f.HeaderComment(t.info.FileHeader)
🟩 78     
🟩 79     	f.Comment(ServiceLoggingMiddlewareName + " writes params, results and working time of method call to provided logger after its execution.").
🟩 80     		Line().Func().Id(ServiceLoggingMiddlewareName).Params(Id(_logger_).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
🟩 81     		Block(t.newLoggingBody(t.info.Iface))
🟩 82     
🟩 83     	f.Line()
🟩 84     
🟩 85     	// Render type logger
🟩 86     	f.Type().Id(serviceLoggingStructName).Struct(
🟩 87     		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
🟩 88     		Id(_next_).Qual(t.info.SourcePackageImport, t.info.Iface.Name),
🟩 89     	)
🟩 90     
🟩 91     	// Render functions
🟩 92     	for _, signature := range t.info.Iface.Methods {
🟩 93     		f.Line()
🟩 94     		f.Add(t.loggingFunc(ctx, signature)).Line()
🟩 95     	}
🟩 96     	if len(t.info.Iface.Methods) > 0 {
🟩 97     		f.Type().Op("(")
🟩 98     	}
🟩 99     	for _, signature := range t.info.Iface.Methods {
🟩 100    		if params := RemoveContextIfFirst(signature.Args); t.calcParamAmount(signature.Name, params) > 0 {
🟩 101    			f.Add(t.loggingEntity(ctx, "log"+requestStructName(signature), signature, params))
🟩 102    		}
🟩 103    		if params := removeErrorIfLast(signature.Results); t.calcParamAmount(signature.Name, params) > 0 {
🟩 104    			f.Add(t.loggingEntity(ctx, "log"+responseStructName(signature), signature, params))
🟩 105    		}
🟩 106    	}
🟩 107    	if len(t.info.Iface.Methods) > 0 {
🟩 108    		f.Op(")")
🟩 109    	}
🟩 110    
🟩 111    	return f
🟩 112    }
⬜ 113    
```

<br/>

And you're done!

<br/>

<!-- THIS IS AN AUTOGENERATED SECTION. DO NOT EDIT THIS SECTION DIRECTLY -->
### Swimm Note

<span id="f-Z1tc5zk">ChooseStrategy</span>[^](#Z1tc5zk) - "generator/template/template.go" L16
```go
	ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error)
```

<span id="f-ZRpVyT">DefaultPath</span>[^](#ZRpVyT) - "generator/template/template.go" L14
```go
	DefaultPath() string
```

<span id="f-ZROtS5">ignoreParams</span>[^](#ZROtS5) - "generator/template/service_logging.go" L119
```go
	t.ignoreParams = make(map[string][]string)
```

<span id="f-Z1eqtoA">NewCreateFileStrategy</span>[^](#Z1eqtoA) - "generator/write_strategy/file.go" L85
```go
func NewCreateFileStrategy(absPath, relPath string) Strategy {
```

<span id="f-2n3btL">Prepare</span>[^](#2n3btL) - "generator/template/template.go" L12
```go
	Prepare(ctx context.Context) error
```

<span id="f-Z2dN3kR">Render</span>[^](#Z2dN3kR) - "generator/template/template.go" L18
```go
	Render(ctx context.Context) write_strategy.Renderer
```

<span id="f-1ghvhI">Template</span>[^](#1ghvhI) - "generator/template/template.go" L9
```go
type Template interface {
```

<br/>

This file was generated by Swimm. [Click here to view it in the app](https://app.swimm.io/repos/Z2l0aHViJTNBJTNBbWljcm9nZW4lM0ElM0FSZWNvTGFicw==/docs/split).