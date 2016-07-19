package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

// DecodeQueryRequest decodes the request from the provided HTTP request, simply
// by JSON decoding from the request body. It's designed to be used in
// transport/http.Server.
func DecodeQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request QueryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return &request, err
}

// EncodeQueryResponse encodes the response to the provided HTTP response
// writer, simply by JSON encoding to the writer. It's designed to be used in
// transport/http.Server.
func EncodeQueryResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// EncodeQueryRequest encodes the request to the provided HTTP request, simply
// by JSON encoding to the request body. It's designed to be used in
// transport/http.Client.
func EncodeQueryRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// DecodeQueryResponse decodes the response from the provided HTTP response,
// simply by JSON decoding from the response body. It's designed to be used in
// transport/http.Client.
func DecodeQueryResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response QueryResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}
