package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mstrings "github.com/dreamsxin/go-kitcli/generator/strings"
	"github.com/dreamsxin/go-kitcli/generator/template"
	lg "github.com/dreamsxin/go-kitcli/logger"
	"github.com/vetcher/go-astra/types"
)

const (
	TagMark         = template.TagMark
	MicrogenMainTag = template.MicrogenMainTag
	ProtobufTag     = "protobuf"
	GRPCClientAddr  = "grpc-addr"

	MiddlewareTag             = template.MiddlewareTag
	LoggingMiddlewareTag      = template.LoggingMiddlewareTag
	RecoveringMiddlewareTag   = template.RecoveringMiddlewareTag
	HttpTag                   = template.HttpTag
	HttpServerTag             = template.HttpServerTag
	HttpClientTag             = template.HttpClientTag
	GrpcTag                   = template.GrpcTag
	GrpcServerTag             = template.GrpcServerTag
	GrpcClientTag             = template.GrpcClientTag
	MainTag                   = template.MainTag
	ErrorLoggingMiddlewareTag = template.ErrorLoggingMiddlewareTag
	TracingMiddlewareTag      = template.TracingMiddlewareTag
	CachingMiddlewareTag      = template.CachingMiddlewareTag
	JSONRPCTag                = template.JSONRPCTag
	JSONRPCServerTag          = template.JSONRPCServerTag
	JSONRPCClientTag          = template.JSONRPCClientTag
	Transport                 = template.Transport
	TransportClient           = template.TransportClient
	TransportServer           = template.TransportServer
	MetricsMiddlewareTag      = template.MetricsMiddlewareTag
	ServiceDiscoveryTag       = template.ServiceDiscoveryTag

	HttpMethodTag  = template.HttpMethodTag
	HttpMethodPath = template.HttpMethodPath
)

func ListTemplatesForGen(ctx context.Context, iface *types.Interface, absOutPath, sourcePath, packageName string, genProto string, genMain bool) (units []*GenerationUnit, err error) {

	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}

	allowedMethods := make(map[string]bool, len(iface.Methods))
	oneToManyStreamMethods := make(map[string]bool, len(iface.Methods))
	manyToManyStreamMethods := make(map[string]bool, len(iface.Methods))
	manyToOneStreamMethods := make(map[string]bool, len(iface.Methods))
	for _, fn := range iface.Methods {
		allowedMethods[fn.Name] = !mstrings.ContainTag(mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag), "-")
		oneToManyStreamMethods[fn.Name] = mstrings.ContainTag(mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag), "one-to-many")
		manyToManyStreamMethods[fn.Name] = mstrings.ContainTag(mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag), "many-to-many")
		manyToOneStreamMethods[fn.Name] = mstrings.ContainTag(mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag), "many-to-one")
	}
	info := &template.GenerationInfo{
		SourcePackageImport:     packageName,
		SourceFilePath:          absSourcePath,
		Iface:                   iface,
		OutputPackageImport:     packageName,
		OutputFilePath:          absOutPath,
		ProtobufPackageImport:   mstrings.FetchMetaInfo(TagMark+ProtobufTag, iface.Docs),
		FileHeader:              defaultFileHeader,
		AllowedMethods:          allowedMethods,
		OneToManyStreamMethods:  oneToManyStreamMethods,
		ManyToManyStreamMethods: manyToManyStreamMethods,
		ManyToOneStreamMethods:  manyToOneStreamMethods,
		ProtobufClientAddr:      mstrings.FetchMetaInfo(TagMark+GRPCClientAddr, iface.Docs),
	}
	lg.Logger.Logln(3, "\nGeneration Info:", info.String())
	/*stubSvc, err := NewGenUnit(ctx, template.NewStubInterfaceTemplate(info), absOutPath)
	if err != nil {
		return nil, err
	}
	units = append(units, stubSvc)*/

	genTags := mstrings.FetchTags(iface.Docs, TagMark+MicrogenMainTag)
	lg.Logger.Logln(2, "Tags:", strings.Join(genTags, ", "))
	uniqueTemplate := make(map[string]template.Template)
	for _, tag := range genTags {
		templates := tagToTemplate(tag, info)
		if templates == nil {
			lg.Logger.Logln(1, "Warning: Unexpected tag", tag)
			continue
		}
		for _, t := range templates {
			uniqueTemplate[t.DefaultPath()] = t
		}
	}
	for _, t := range uniqueTemplate {
		unit, err := NewGenUnit(ctx, t, absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, unit)
	}
	if genProto != "" {
		u, err := NewGenUnit(ctx, template.NewProtoTemplate(info, genProto), absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, u)
	}
	if genMain {
		u, err := NewGenUnit(ctx, template.NewMainTemplate(info), absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, u)
	}
	return units, nil
}

func tagToTemplate(tag string, info *template.GenerationInfo) (tmpls []template.Template) {
	switch tag {
	case MiddlewareTag:
		return append(tmpls, template.NewMiddlewareTemplate(info))
	case LoggingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewLoggingTemplate(info),
		)
	case GrpcTag:
		return append(
			append(tmpls, tagToTemplate(Transport, info)...),
			template.NewGRPCClientTemplate(info),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcClientTag:
		return append(
			append(tmpls, tagToTemplate(TransportClient, info)...),
			template.NewGRPCClientTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcServerTag:
		return append(
			append(tmpls, tagToTemplate(TransportServer, info)...),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case HttpTag:
		return append(
			append(tmpls, tagToTemplate(Transport, info)...),
			template.NewHttpServerTemplate(info),
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpServerTag:
		return append(
			append(tmpls, tagToTemplate(TransportServer, info)...),
			template.NewHttpServerTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpClientTag:
		return append(
			append(tmpls, tagToTemplate(TransportClient, info)...),
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case RecoveringMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewRecoverTemplate(info),
		)
	case MainTag:
		lg.Logger.Logln(1, "Warning: Tag main is deprecated, use flag -main instead.")
		return nil
	case ErrorLoggingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewErrorLoggingTemplate(info),
		)
	case CachingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewCacheMiddlewareTemplate(info),
		)
	case TracingMiddlewareTag:
		return append(tmpls, template.EmptyTemplate{})
	case MetricsMiddlewareTag:
		return append(tmpls, template.EmptyTemplate{})
	case ServiceDiscoveryTag:
		return append(tmpls, template.EmptyTemplate{})
	case Transport:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsClientTemplate(info),
			template.NewEndpointsServerTemplate(info),
		)
	case TransportClient:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsClientTemplate(info),
		)
	case TransportServer:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsServerTemplate(info),
		)
	}
	return nil
}

func resolvePackagePath(outPath string) (string, error) {
	lg.Logger.Logln(3, "Try to resolve path for", outPath, "package...")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	lg.Logger.Logln(4, "GOPATH:", gopath)

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}
	lg.Logger.Logln(4, "Resolving path:", absOutPath)

	for _, path := range strings.Split(gopath, ":") {
		gopathSrc := filepath.Join(path, "src")
		if strings.HasPrefix(absOutPath, gopathSrc) {
			return absOutPath[len(gopathSrc)+1:], nil
		}
	}
	return "", fmt.Errorf("path(%s) not in GOPATH(%s)", absOutPath, gopath)
}
