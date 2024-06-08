package functions

import (
	"fmt"
)

func must(v any) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("missing")
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return nil, fmt.Errorf("missing")
		}
	}
	return v, nil
}
