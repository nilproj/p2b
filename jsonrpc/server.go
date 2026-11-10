package jsonrpc

import (
	"encoding/json"
	"io"
)

type Server struct {
	decoder *json.Decoder
	encoder *json.Encoder

	handlers map[string]Handler
}

func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{
		decoder:  json.NewDecoder(r),
		encoder:  json.NewEncoder(w),
		handlers: map[string]Handler{},
	}
}

func (srv *Server) Handle(method string, hdr Handler) {
	srv.handlers[method] = hdr
}

func (srv *Server) ListenAndServe() error {
	for {
		var req jrpcRequest
		if err := srv.decoder.Decode(&req); err != nil {
			return err
		}

		w := Writer{srv.encoder, req}
		r := &Request{Params: req.Params}
		srv.handlers[req.Method](w, r)
	}
}
