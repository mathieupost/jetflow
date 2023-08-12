package jetflow

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Request struct {
	OperationID string
	RequestID   string
	ClientID    string

	Name   string
	ID     string
	Method string
	Args   []byte
}

type Response struct {
	RequestID string

	Values []byte
	Error  error
}

// UnmarshalJSON implements json.Unmarshaler
func (r *Response) UnmarshalJSON(data []byte) error {
	var res struct {
		RequestID string

		Values []byte
		Error  string
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return errors.Wrap(err, "unmarshalling Reply")
	}

	r.Values = res.Values
	if res.Error != "" {
		r.Error = errors.New(res.Error)
	}
	return nil
}
