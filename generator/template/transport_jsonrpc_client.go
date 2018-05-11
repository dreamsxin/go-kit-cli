package template

import (
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/go-astra/types"
)

type jsonrpcClientTemplate struct {
	Info     *GenerationInfo
	tracing  bool
	prefixes map[string]string
	suffixes map[string]string
}

func NewJSONRPCClientTemplate(info *GenerationInfo) Template {
	return &jsonrpcClientTemplate{
		Info: info.Copy(),
	}
}

func (t *jsonrpcClientTemplate) DefaultPath() string {
	return "./transport/jsonrpc/client.go"
}

func (t *jsonrpcClientTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
}

func (t *jsonrpcClientTemplate) Prepare() error {
	tags := util.FetchTags(t.Info.Iface.Docs, TagMark+MicrogenMainTag)
	for _, tag := range tags {
		switch tag {
		case TracingTag:
			t.tracing = true
		}
	}
	t.prefixes = make(map[string]string)
	t.suffixes = make(map[string]string)
	for _, fn := range t.Info.Iface.Methods {
		if s := util.FetchTags(fn.Docs, TagMark+prefixJSONRPCAnnotationTag); len(s) > 0 {
			t.prefixes[fn.Name] = s[0]
		}
		if s := util.FetchTags(fn.Docs, TagMark+suffixJSONRPCAnnotationTag); len(s) > 0 {
			t.suffixes[fn.Name] = s[0]
		}
	}
	return nil
}

func (t *jsonrpcClientTemplate) Render() write_strategy.Renderer {
	f := NewFile("transportjsonrpc")
	f.ImportAlias(t.Info.SourcePackageImport, serviceAlias)
	f.PackageComment(t.Info.FileHeader)
	f.PackageComment(`DO NOT EDIT.`)

	f.Func().Id("NewJSONRPCClient").ParamsFunc(func(p *Group) {
		p.Id("addr").Id("string")
		if t.tracing {
			p.Id("logger").Qual(PackagePathGoKitLog, "Logger")
		}
		if t.tracing {
			p.Id("tracer").Qual(PackagePathOpenTracingGo, "Tracer")
		}
		p.Id("opts").Op("...").Qual(PackagePathGoKitTransportJSONRPC, "ClientOption")
	}).Params(
		Qual(t.Info.SourcePackageImport, t.Info.Iface.Name),
		Error(),
	).Block(
		t.clientBody(),
	)

	return f
}

func (t *jsonrpcClientTemplate) clientBody() *Statement {
	g := &Statement{}
	g.If(
		Op("!").Qual(PackagePathStrings, "HasPrefix").Call(Id("addr"), Lit("http")),
	).Block(
		Id("addr").Op("=").Lit("http://").Op("+").Id("addr"),
	)
	g.Line().List(Id("u"), Err()).Op(":=").Qual(PackagePathUrl, "Parse").Call(Id("addr"))
	g.Line().If(Err().Op("!=").Nil()).Block(
		Return(Nil(), Err()),
	)
	if t.tracing {
		g.Line().Id("opts").Op("=").Append(Id("opts"), Qual(PackagePathGoKitTransportJSONRPC, "ClientBefore").Call(
			Line().Qual(PackagePathGoKitTracing, "ContextToHTTP").Call(Id("tracer"), Id("logger")).Op(",").Line(),
		))
	}
	g.Line().Return(Op("&").Qual(t.Info.SourcePackageImport, "Endpoints").Values(DictFunc(
		func(d Dict) {
			for _, fn := range t.Info.Iface.Methods {
				client := &Statement{}
				if t.tracing {
					client.Qual(PackagePathGoKitTracing, "TraceClient").Call(
						Line().Id("tracer"),
						Line().Lit(fn.Name),
						Line(),
					).Op("(").Line()
					defer func() { client.Op(",").Line().Op(")") }() // defer in for loop is OK
				}
				client.Qual(PackagePathGoKitTransportJSONRPC, "NewClient").Call(
					Line().Id("u"),
					Line().Lit(t.prefixes[fn.Name]+fn.Name+t.suffixes[fn.Name]),
					Line().Append(
						Line().Add(t.clientOpts(fn)),
						Line().Qual(PackagePathGoKitTransportJSONRPC, "ClientRequestEncoder").
							Call(Qual(pathToJSONRPCConverter(t.Info.SourcePackageImport), encodeRequestName(fn))),
						Line().Qual(PackagePathGoKitTransportJSONRPC, "ClientResponseDecoder").
							Call(Qual(pathToJSONRPCConverter(t.Info.SourcePackageImport), decodeResponseName(fn))).Op(",").Line(),
					).Op("...").Line(),
				).Dot("Endpoint").Call()
				d[Id(endpointStructName(fn.Name))] = client
			}
		},
	)), Nil())
	return g
}

func (t *jsonrpcClientTemplate) clientOpts(fn *types.Function) *Statement {
	s := &Statement{}
	s.Id("opts")
	return s
}

func pathToJSONRPCConverter(servicePath string) string {
	return filepath.Join(servicePath, "transport/converter/jsonrpc")
}
