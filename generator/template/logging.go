package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const (
	loggerVarName            = "logger"
	nextVarName              = "next"
	serviceLoggingStructName = "serviceLogging"
)

type LoggingTemplate struct {
	PackagePath string
}

// Render all logging.go file.
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package stringsvc
//
//		import (
//			context "context"
//			svc "github.com/devimteam/microgen/test/svc"
//			log "github.com/go-kit/kit/log"
//			time "time"
//		)
//
//		func ServiceLogging(logger log.Logger) Middleware {
//			return func(next svc.StringService) svc.StringService {
//				return stringServiceLoggingMiddleware{
//					logger: logger,
//					next:   next,
//				}
//			}
//		}
//
//		type stringServiceLoggingMiddleware struct {
//			logger log.Logger
//			next   svc.StringService
//		}
//
//		func (s *stringServiceLoggingMiddleware) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text, "symbol", symbol,
//					"count", count, "positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t LoggingTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Func().Id(util.ToUpperFirst(serviceLoggingStructName)).Params(Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newLoggingBody(i))

	f.Line()

	// Render type logger
	f.Type().Id(serviceLoggingStructName).Struct(
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.PackagePath, i.Name),
	)

	// Render functions
	for _, signature := range i.FuncSignatures {
		f.Line()
		f.Add(loggingFunc(signature, i))
	}

	return f
}

func (LoggingTemplate) Path() string {
	return "./middleware/logging.go"
}

// Render body for new logging middleware.
//
//		return func(next svc.StringService) svc.StringService {
//			return StringServiceLoggingMiddleware{
//				logger: logger,
//				next:   next,
//			}
//		}
//
func (t LoggingTemplate) newLoggingBody(i *parser.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.PackagePath, i.Name),
	).Params(
		Qual(t.PackagePath, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceLoggingStructName).Values(
			Dict{
				Id(loggerVarName): Id(loggerVarName),
				Id(nextVarName):   Id(nextVarName),
			},
		))
	}))
}

// Render logging middleware for interface method.
//
//		func (s *StringServiceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log("method", "Count",
//					"text", text, "symbol", symbol,
//					"count", count, "positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func loggingFunc(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	return methodDefinition(serviceLoggingStructName, signature).
		BlockFunc(loggingFuncBody(signature))
}

// Render logging function body with request/response and time tracking.
//
//		defer func(begin time.Time) {
//			s.logger.Log("method", "Count",
//				"text", text, "symbol", symbol,
//				"count", count, "positions", positions,
//				"took", time.Since(begin))
//		}(time.Now())
//		return s.next.Count(ctx, text, symbol)
//
func loggingFuncBody(signature *parser.FuncSignature) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params(Id("begin").Qual(PackagePathTime, "Time")).Block(
			Id(util.FirstLowerChar(serviceLoggingStructName)).Dot(loggerVarName).Dot("Log").Call(
				Line().Lit("method"), Lit(signature.Name),
				Line().Add(paramsNameAndValue(signature.Params)),
				Line().Add(paramsNameAndValue(signature.Results)),
				Line().Lit("took"), Qual(PackagePathTime, "Since").Call(Id("begin")),
			),
		).Call(Qual(PackagePathTime, "Now").Call())

		g.Return().Id(util.FirstLowerChar(serviceLoggingStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(signature.Params))
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err, "result", result,
//
func paramsNameAndValue(fields []*parser.FuncField) *Statement {
	return ListFunc(func(g *Group) {
		for i, field := range fields {
			if i == 0 && isContext(field) {
				continue
			}
			g.List(Lit(field.Name), Id(field.Name))
		}
	})
}
