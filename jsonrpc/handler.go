package jsonrpc

import "encoding/json"

type Handler func(Writer, *Request)

type Request struct {
	Params []byte
}

type Writer struct {
	encoder *json.Encoder
	req     jrpcRequest
}

func (w Writer) Write(v any) error {
	return w.encoder.Encode(jrpcResponse{
		JSONRPC: version,
		ID:      w.req.ID,
		Result:  v,
	})
}
