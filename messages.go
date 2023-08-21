package jetflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Method string

const (
	MethodPrepare  Method = "__PREPARE__"
	MethodCommit   Method = "__COMMIT__"
	MethodRollback Method = "__ROLLBACK__"
)

type Request struct {
	OperationID string `json:"o"`
	RequestID   string `json:"r"`

	Name   string `json:"n"`
	ID     string `json:"i"`
	Method string `json:"m"`
	Args   []byte `json:"a"`
}

// String returns a string representation of the request.
func (r *Request) String() string {
	return fmt.Sprintf("%s %s %s(%s).%s(%s)",
		r.OperationID, r.RequestID,
		r.Name, r.ID, r.Method, string(r.Args),
	)
}

func (r *Request) Response(ctx context.Context, values []byte, err error) *Response {
	return &Response{
		RequestID:         r.RequestID,
		InvolvedOperators: InvolvedOperatorsFromContext(ctx),
		Values:            values,
		Error:             err,
	}
}

type Response struct {
	RequestID         string
	InvolvedOperators map[string]map[string]bool

	Values []byte
	Error  error
}

type jsonResponse struct {
	RequestID         string                     `json:"r"`
	InvolvedOperators map[string]map[string]bool `json:"o"`

	Values []byte `json:"v"`
	Error  string `json:"e"`
}

func (r Response) MarshalJSON() ([]byte, error) {
	var rerr string
	if r.Error != nil {
		rerr = r.Error.Error()
	}

	res := jsonResponse{
		r.RequestID,
		r.InvolvedOperators,
		r.Values,
		rerr,
	}

	data, err := json.Marshal(res)
	return data, errors.Wrap(err, "marshal Result")
}

// UnmarshalJSON implements json.Unmarshaler
func (r *Response) UnmarshalJSON(data []byte) error {
	var res jsonResponse

	err := json.Unmarshal(data, &res)
	if err != nil {
		return errors.Wrap(err, "unmarshalling Reply")
	}

	var rerr error
	if res.Error != "" {
		rerr = errors.New(res.Error)
	}

	*r = Response{
		res.RequestID,
		res.InvolvedOperators,
		res.Values,
		rerr,
	}
	return nil
}
