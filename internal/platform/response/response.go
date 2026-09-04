package response

import (
	"encoding/json"
	"errors"
	"io"
)

type Envelope struct {
	Data any `json:"data,omitempty"`
}

type Message struct {
	Message string `json:"message"`
}

type ErrorBody struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func DecodeJSON(body []byte, destination any) error {
	decoder := json.NewDecoder(io.LimitReader(newByteReader(body), int64(len(body)+1)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func Encode(value any) ([]byte, error) { return json.Marshal(value) }

type byteReader struct {
	data []byte
	read int
}

func newByteReader(data []byte) *byteReader { return &byteReader{data: data} }

func (r *byteReader) Read(p []byte) (int, error) {
	if r.read >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.read:])
	r.read += n
	return n, nil
}
