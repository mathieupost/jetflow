package jetflow

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Result struct {
	Error  error
	Values []byte
}

type jsonResult struct {
	Error  string
	Values []byte
}

func (r *Result) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	var res jsonResult

	err := json.Unmarshal(data, &res)
	if err != nil {
		return errors.Wrap(err, "unmarshal Result")
	}

	r.Values = res.Values
	if res.Error != "" {
		r.Error = errors.New(res.Error)
	}
	return nil
}

func (r Result) MarshalJSON() ([]byte, error) {
	var res jsonResult
	res.Values = r.Values

	if r.Error != nil {
		res.Error = r.Error.Error()
	}

	data, err := json.Marshal(res)
	return data, errors.Wrap(err, "marshal Result")
}
