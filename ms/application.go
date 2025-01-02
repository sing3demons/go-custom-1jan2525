package ms

import (
	"log"
	"net/http"
)

type IApplication interface {
	Get(path string, handler HandleFunc, middlewares ...Middleware)
	Post(path string, handler HandleFunc, middlewares ...Middleware)
	Use(middlewares ...Middleware)
	Start(port string)
}

type muxApplication struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewApplication() IApplication {
	app := http.NewServeMux()

	return &muxApplication{
		mux: app,
	}
}

func (app *muxApplication) Use(middlewares ...Middleware) {
	app.middlewares = append(app.middlewares, middlewares...)
}

func (app *muxApplication) Get(path string, handler HandleFunc, middlewares ...Middleware) {
	app.mux.HandleFunc(http.MethodGet+" "+path, func(w http.ResponseWriter, r *http.Request) {
		r = setParam(path, r)
		preHandle(handler, preMiddleware(app, middlewares)...)(newMuxContext(w, r))
	})
}

func preMiddleware(app *muxApplication, middlewares []Middleware) []Middleware {
	var m []Middleware
	if len(app.middlewares) > 0 {
		m = append(m, app.middlewares...)
	}

	if len(middlewares) > 0 {
		m = append(m, middlewares...)
	}
	return m
}

func (app *muxApplication) Post(path string, handler HandleFunc, middlewares ...Middleware) {
	app.mux.HandleFunc(http.MethodPost+" "+path, func(w http.ResponseWriter, r *http.Request) {
		preHandle(handler, preMiddleware(app, middlewares)...)(newMuxContext(w, r))
	})
}

func (app *muxApplication) Start(port string) {
	server := http.Server{
		Handler: app.mux,
		Addr:    ":" + port,
		// ReadTimeout:  timeout,
		// WriteTimeout: timeout,
	}

	log.Printf("Start server: %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
