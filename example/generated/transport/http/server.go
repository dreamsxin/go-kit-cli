// This file was automatically generated by "microgen 0.7.0b" utility.
// Please, do not edit.
package transporthttp

import (
	generated "github.com/devimteam/microgen/example/generated"
	http2 "github.com/devimteam/microgen/example/generated/transport/converter/http"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	http1 "net/http"
)

func NewHTTPHandler(endpoints *generated.Endpoints, opts ...http.ServerOption) http1.Handler {
	mux := mux.NewRouter()
	mux.Methods("GET").Path("uppercase/{strings-map}").Handler(
		http.NewServer(
			endpoints.UppercaseEndpoint,
			http2.DecodeHTTPUppercaseRequest,
			http2.EncodeHTTPUppercaseResponse,
			opts...))
	mux.Methods("GET").Path("count").Handler(
		http.NewServer(
			endpoints.CountEndpoint,
			http2.DecodeHTTPCountRequest,
			http2.EncodeHTTPCountResponse,
			opts...))
	mux.Methods("POST").Path("test-case").Handler(
		http.NewServer(
			endpoints.TestCaseEndpoint,
			http2.DecodeHTTPTestCaseRequest,
			http2.EncodeHTTPTestCaseResponse,
			opts...))
	return mux
}
