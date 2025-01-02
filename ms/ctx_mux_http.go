package ms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpContext struct {
	w http.ResponseWriter
	r *http.Request
}

func newMuxContext(w http.ResponseWriter, r *http.Request) IContext {
	return &HttpContext{
		w: w,
		r: r,
	}
}

type HandleFunc func(ctx IContext) error

type Middleware func(HandleFunc) HandleFunc

func (c *HttpContext) Log(message string) {
	fmt.Println("Context:", message)
}

func (c *HttpContext) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *HttpContext) Param(name string) string {
	v := c.r.Context().Value(ContextKey(name))
	var value string
	switch v := v.(type) {
	case string:
		value = v
	}
	c.r = c.r.WithContext(context.WithValue(c.r.Context(), ContextKey(name), nil))
	return value
}

func (c *HttpContext) ReadInput(data interface{}) error {
	return nil
}

func (c *HttpContext) Response(responseCode int, responseData interface{}) {
	json.NewEncoder(c.w).Encode(responseData)
}

func preHandle(final HandleFunc, middlewares ...Middleware) HandleFunc {
	if final == nil {
		panic("no final handler")
		// Or return a default handler.
	}
	// Execute the middleware in the same order and return the final func.
	// This is a confusing and tricky construct :)
	// We need to use the reverse order since we are chaining inwards.
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final) // mw1(mw2(mw3(final)))
	}
	return final
}
