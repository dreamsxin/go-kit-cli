// Code generated by microgen 0.9.0. DO NOT EDIT.

// DO NOT EDIT.
package transportgrpc

import (
	log "github.com/go-kit/kit/log"
	opentracing "github.com/go-kit/kit/tracing/opentracing"
	grpc "github.com/go-kit/kit/transport/grpc"
	empty "google.golang.org/protobuf/types/known/emptypb"
	opentracinggo "github.com/opentracing/opentracing-go"
	transport "github.com/dreamsxin/go-kitcli/examples/generated/transport"
	pb "github.com/dreamsxin/go-kitcli/examples/protobuf"
	context "golang.org/x/net/context"
)

type stringServiceServer struct {
	uppercase   grpc.Handler
	count       grpc.Handler
	testCase    grpc.Handler
	dummyMethod grpc.Handler
}

func NewGRPCServer(endpoints *transport.EndpointsSet, logger log.Logger, tracer opentracinggo.Tracer, opts ...grpc.ServerOption) pb.StringServiceServer {
	return &stringServiceServer{
		count: grpc.NewServer(
			endpoints.CountEndpoint,
			_Decode_Count_Request,
			_Encode_Count_Response,
			append(opts, grpc.ServerBefore(
				opentracing.GRPCToContext(tracer, "Count", logger)))...,
		),
		dummyMethod: grpc.NewServer(
			endpoints.DummyMethodEndpoint,
			_Decode_DummyMethod_Request,
			_Encode_DummyMethod_Response,
			append(opts, grpc.ServerBefore(
				opentracing.GRPCToContext(tracer, "DummyMethod", logger)))...,
		),
		testCase: grpc.NewServer(
			endpoints.TestCaseEndpoint,
			_Decode_TestCase_Request,
			_Encode_TestCase_Response,
			append(opts, grpc.ServerBefore(
				opentracing.GRPCToContext(tracer, "TestCase", logger)))...,
		),
		uppercase: grpc.NewServer(
			endpoints.UppercaseEndpoint,
			_Decode_Uppercase_Request,
			_Encode_Uppercase_Response,
			append(opts, grpc.ServerBefore(
				opentracing.GRPCToContext(tracer, "Uppercase", logger)))...,
		),
	}
}

func (S *stringServiceServer) Uppercase(ctx context.Context, req *pb.UppercaseRequest) (*pb.UppercaseResponse, error) {
	_, resp, err := S.uppercase.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UppercaseResponse), nil
}

func (S *stringServiceServer) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	_, resp, err := S.count.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CountResponse), nil
}

func (S *stringServiceServer) TestCase(ctx context.Context, req *pb.TestCaseRequest) (*pb.TestCaseResponse, error) {
	_, resp, err := S.testCase.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.TestCaseResponse), nil
}

func (S *stringServiceServer) DummyMethod(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	_, resp, err := S.dummyMethod.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*empty.Empty), nil
}
