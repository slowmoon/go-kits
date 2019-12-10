package errors

import (
	"fmt"
	"strings"
)

type MultiError []error

func (m MultiError) Error() string {
	prefix := "multi error happen"
	var errstr  []string
	for i, err := range m {
		errstr = append(errstr, fmt.Sprintf("%d: %s", i, err.Error()))
	}
	joined := strings.Join(errstr, ",")
	return  fmt.Sprintf("[%s] %s", prefix, joined)
}

func (m *MultiError)Add(err error)  {
	*m = append(*m, err)
}