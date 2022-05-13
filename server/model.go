package server

import (
	"encoding/json"
	"errors"
	"io"
)

var (
	ErrMissingSubject = errors.New("Missing subject")
)

type Request struct {
	Subject string `json:"subject"`
	Payload any    `json:"payload"`
}

type Response struct {
	Payload any    `json:"payload"`
	Error   string `json:"error"`
}

func isValidRequest(Request Request) error {
	if Request.Subject == "" {
		return ErrMissingSubject
	}
	return nil
}

func ConvertHTTPRequestToNats(reader io.Reader) (*Request, error) {
	var req Request
	err := json.NewDecoder(reader).Decode(&req)
	if err != nil {
		return nil, err
	}

	err = isValidRequest(req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func ConvertWSMsgToNats(b []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(b, &req)
	if err != nil {
		return nil, err
	}
	err = isValidRequest(req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func ConvertNatsResponseToWSMsg(req any) ([]byte, error) {
	var res Response
	res.Payload = req
	return json.Marshal(res)
}
