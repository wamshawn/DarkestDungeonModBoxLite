package failure

import (
	"encoding/json"
)

type Failure struct {
	Title       string `json:"error"`
	Description string `json:"description"`
}

type Failures []Failure

func (f Failures) Error() string {
	b, _ := json.Marshal(f)
	return string(b)
}

func (f Failures) Append(title, description string) Failures {
	return append(f, Failure{Title: title, Description: description})
}

func (f Failures) Wrap(err error) Failures {
	if err == nil {
		return f
	}
	switch e := err.(type) {
	case Failures:
		return append(f, e...)
	default:
		return f.Append("错误", e.Error())
	}
}

func Failed(title, description string) Failures {
	ff := make(Failures, 0, 1)
	ff = append(ff, Failure{Title: title, Description: description})
	return ff
}
