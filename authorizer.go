package brute

import (
	"net/http"
	"path/filepath"
	"context"
	"fmt"
)

var authorizeHandler *ControllerEndpoint

type auth struct {
	next 				http.HandlerFunc
	unauthorizedPage 	http.HandlerFunc
	protected 			bool
}

type AuthorizeHandler interface {
	Success(handler http.HandlerFunc) AuthorizeHandler
	Failed(handler http.HandlerFunc) AuthorizeHandler
	Handler() http.HandlerFunc
}

func LoadAuthorizer(route Route) AuthorizeHandler {
	if authorizeHandler == nil {
		authorizeHandler = &ControllerEndpoint{
				ProjectName: projectName,
				Route: route,
				runtimeFile: filepath.Join(cwd, "bin", "endpoints", route.Directory),
			}
	}

	if route.RouteConfig != nil && route.Protected {
		return &auth{protected: true}
	}

	return &auth{}
}

func (auth *auth) Success(handler http.HandlerFunc) AuthorizeHandler {
	if handler == nil {
		panic("this handler must be supplied") //This shouldn't happen
	}

	auth.next = handler

	return auth
}

func (auth *auth) Failed(handler http.HandlerFunc) AuthorizeHandler {
	auth.unauthorizedPage = handler
	if handler == nil {
		auth.unauthorizedPage = defaultUnauthorizedHandler
	}

	return auth
}

func (auth *auth) Handler() http.HandlerFunc {
	if !auth.protected {
		return auth.next
	}

	if auth.unauthorizedPage == nil {
		auth.unauthorizedPage = defaultUnauthorizedHandler
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authContext := &MiddlewareWriterContext{header: r.Header}
		authorizeHandler.ServeHTTP(authContext, r) //This blocks until remote endpoint finishes execution
		if authContext.httpCode != 403 {
			r.WithContext(context.WithValue(r.Context(), "policy", authContext))
			auth.next(w, r)
		} else {
			auth.unauthorizedPage(w, r)
		}
	}
}

type MiddlewareWriterContext struct {
	httpCode int
	buffer []byte
	header http.Header
}

func (context *MiddlewareWriterContext) Header() http.Header {
	return context.header
}

func (context *MiddlewareWriterContext) Write(data []byte) (int, error) {
	context.buffer = append(context.buffer, data...)
	return len(context.buffer), nil
}

func (context *MiddlewareWriterContext) WriteHeader(statusCode int) {
	context.httpCode = statusCode
}