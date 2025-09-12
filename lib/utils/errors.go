package utils

import (
	"fmt"
	"strings"
)

func Stringify(err error) string {
	if err == nil {
		return ""
	}
	if err, ok := err.(interface {
		Unwrap() []error
	}); ok {
		return fmt.Sprintf("%s%s%s", "[",
			strings.Join(Map(
				err.Unwrap(),
				func(err error) string { return Stringify(err) },
			), ";"),
			"]")
	}
	return err.Error()
}
