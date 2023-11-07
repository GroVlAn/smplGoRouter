package smplGoRouter

import (
	"net/http"
	"sync"
)

type Handler func(http.ResponseWriter, *http.Request)

type MiddlewareAdder interface {
	AddMiddleware(middleware Handler)
}

type Getter interface {
	Get(path string, handler Handler)
}

type Poster interface {
	Post(path string, handler Handler)
}

type Putter interface {
	Put(path string, handler Handler)
}

type Deleter interface {
	Delete(path string, handler Handler)
}

type CRUD interface {
	Getter
	Poster
	Putter
	Deleter
}

type HandleAdder interface {
	Handle(path string, handler http.Handler)
}

type HandleFuncAdder interface {
	HandleFunc(path string, handler http.HandlerFunc)
}

type Router struct {
	Mux         *http.ServeMux
	middlewares []Handler
	Addr        string
	MiddlewareAdder
	CRUD
	HandleAdder
	HandleFuncAdder
}

func NewRouter(addr string) *Router {
	return &Router{
		Mux:         http.NewServeMux(),
		middlewares: make([]Handler, 0),
		Addr:        addr,
	}
}

func (r *Router) AddMiddleware(middleware Handler) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) middlewareWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var wg sync.WaitGroup

		for _, f := range r.middlewares {
			wg.Add(1)
			go func(f Handler) {
				defer wg.Done()
				f(w, req)
			}(f)
		}
		wg.Wait()
		next.ServeHTTP(w, req)
	})
}

func (r *Router) handle(path string, handler Handler, method string) {
	path = r.Addr + path

	handle := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == method {
			handler(w, req)
		}
	}

	r.Mux.Handle(path, r.middlewareWrapper(http.HandlerFunc(handle)))
}

func (r *Router) Get(path string, handler Handler) {
	r.handle(path, handler, http.MethodGet)
}

func (r *Router) Post(path string, handler Handler) {
	r.handle(path, handler, http.MethodPost)
}

func (r *Router) Put(path string, handler Handler) {
	r.handle(path, handler, http.MethodPut)
}

func (r *Router) Delete(path string, handler Handler) {
	r.handle(path, handler, http.MethodDelete)
}

func (r *Router) Handle(path string, handler http.Handler) {
	path = r.Addr + path
	r.Mux.Handle(path, r.middlewareWrapper(handler))
}

func (r *Router) HandleFunc(path string, handler http.HandlerFunc) {
	path = r.Addr + path
	r.Mux.Handle(path, r.middlewareWrapper(handler))
}
