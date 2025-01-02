package ms

import (
	"context"
	"net/http"
	"strings"
)

func removeBraces(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "{", ""), "}", "")
}

type ContextKey string

func setParam(path string, r *http.Request) *http.Request {
	subPath := strings.Split(path, "/")
	sss := strings.Split(r.URL.Path, "/")

	// Remove empty strings from the slice
	if len(subPath) == len(sss) {
		for i := 0; i < len(subPath); i++ {
			var k, v string
			if subPath[i] != "" {
				k = subPath[i]
			}
			if sss[i] != "" {
				v = sss[i]
			}

			if k != "" && v != "" && k != v {
				key := removeBraces(k)
				if key != "" {
					ctx := context.WithValue(r.Context(), ContextKey(key), v)
					r = r.WithContext(ctx)
				}

			}
		}
	}
	return r
}
