package krouter

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// @Author KHighness
// @Update 2022-11-08

var (
	// ErrGenerateParameters is returned when generating a route with invalid parameters.
	ErrGenerateParameters = errors.New("params contains invalid parameters")
	// ErrRouteNotFound is returned when generating a route that can not find route in tree,
	ErrRouteNotFound = errors.New("route not found")
	// ErrMethodNotFound is returned when generating a route that can not find method in tree.
	ErrMethodNotFound = errors.New("method not found")
	// ErrPatternGrammar is returned when generating a route that pattern grammar error.
	ErrPatternGrammar = errors.New("pattern grammar error")
)

var (
	defaultPattern = `[\w]+`
	idPattern      = `[\d]+`
	idKey          = `id`
)

// methods enumerates all the valid http methods.
var methods = map[string]struct{}{
	http.MethodGet:    {},
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodDelete: {},
	http.MethodPatch:  {},
	http.MethodHead:   {},
	// http.MethodOptions: {},
}

// Middleware defines a function which is used for web middleware.
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// Parameters records parameters.
type Parameters struct {
	routeName string
}

// Router is a simple HTTP route multiplexer that parses a request path,
// records any URL params, and executes an end handler.
type Router struct {
	prefix string
	// middleware records the middleware stack
	middleware []Middleware
	// tree routers whose key is method and value is Tree.
	trees      map[string]*Tree
	parameters Parameters
	// notFound is a custom handler for not-found route
	notFound http.HandlerFunc
	// PanicHandler handles panic.
	PanicHandler func(w http.ResponseWriter, r *http.Request, err interface{})
}

// New creates a Router.
func New() *Router {
	return &Router{
		trees: make(map[string]*Tree),
	}
}

// Handle registers a new http.HandlerFunc with the given path and method.
func (r *Router) Handle(method string, path string, handle http.HandlerFunc) {
	if _, ok := methods[method]; !ok {
		panic("invalid method: " + method)
	}

	tree, ok := r.trees[method]
	if !ok {
		tree = NewTree()
		r.trees[method] = tree
	}

	if r.prefix != "" {
		path = r.prefix + "/" + path
	}

	if routeName := r.parameters.routeName; routeName != "" {
		tree.parameters.routeName = routeName
	}

	tree.Register(path, handle, r.middleware...)
}

// Get adds the route `path` which matches a GET http method to execute the `handle` function.
func (r *Router) Get(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodGet, path, handle)
}

// Post adds the route `path` which matches a Post http method to execute the `handle` function.
func (r *Router) Post(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodPost, path, handle)
}

// Put adds the route `path` which matches a Put http method to execute the `handle` function.
func (r *Router) Put(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodPut, path, handle)
}

// Delete adds the route `path` which matches a Delete http method to execute the `handle` function.
func (r *Router) Delete(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodDelete, path, handle)
}

// Head adds the route `path` which matches a Patch http method to execute the `handle` function.
func (r *Router) Patch(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodPatch, path, handle)
}

// Head adds the route `path` which matches a Head http method to execute the `handle` function.
func (r *Router) Head(path string, handle http.HandlerFunc) {
	r.Handle(http.MethodHead, path, handle)
}

// GetAndName is short for Get and Named routeName.
func (r *Router) GetAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Get(path, handle)
}

// PostAndName is short for Post and Named routeName
func (r *Router) PostAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Post(path, handle)
}

// DeleteAndName is short for Delete and Named routeName
func (r *Router) DeleteAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Delete(path, handle)
}

// PutAndName is short for Put and Named routeName
func (r *Router) PutAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Put(path, handle)
}

// PatchAndName is short for Patch and Named routeName
func (r *Router) PatchAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Patch(path, handle)
}

// HEADAndName is short for Head and Named routeName
func (r *Router) HEADAndName(path string, handle http.HandlerFunc, routeName string) {
	r.parameters.routeName = routeName
	r.Head(path, handle)
}

// Group creates a router group if there is a prefix that uses prefix.
func (r *Router) Group(prefix string) *Router {
	return &Router{
		prefix:     prefix,
		trees:      r.trees,
		middleware: r.middleware,
	}
}

// Generate returns reverse routing by method, routeName and params.
func (r *Router) Generate(method string, routeName string, params map[string]string) (string, error) {
	tree, ok := r.trees[method]
	if !ok {
		return "", ErrMethodNotFound
	}

	route, ok := tree.routes[routeName]
	if !ok {
		return "", ErrRouteNotFound
	}

	var segments []string
	list := splitPattern(route.path)
	for _, segment := range list {
		if string(segment[0]) == ":" {
			key := params[string(segment[1:])]
			regex := regexp.MustCompile(defaultPattern)
			if one := regex.Find([]byte(key)); one == nil {
				return "", ErrGenerateParameters
			}
			segments = append(segments, key)
			continue
		}

		if string(segment[0]) == "{" {
			segmentLen := len(segment)
			if string(segment[segmentLen-1]) == "}" {
				splitList := strings.Split(segment[1:segmentLen-1], ":")
				regex := regexp.MustCompile(splitList[1])
				key := params[splitList[0]]
				if one := regex.Find([]byte(key)); one == nil {
					return "", ErrGenerateParameters
				}
				segments = append(segments, key)
				continue
			}
			return "", ErrPatternGrammar
		}

		if string(segment[len(segment)-1]) == "}" && string(segment[0]) != "{" {
			return "", ErrPatternGrammar
		}
		segments = append(segments, segment)
	}

	return "/" + strings.Join(segments, "/"), nil
}

// Use appends middleware handler to the middleware stack.
func (r *Router) Use(middleware ...Middleware) {
	if len(middleware) > 0 {
		r.middleware = append(r.middleware, middleware...)
	}
}

// NotFoundFunc registers a custom handler when the request route is not found.
func (r *Router) NotFoundFunc(handler http.HandlerFunc) {
	r.notFound = handler
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestURL := req.URL.Path

	if r.PanicHandler != nil {
		defer func() {
			if err := recover(); err != nil {
				r.PanicHandler(w, req, err)
			}
		}()
	}

	if _, ok := r.trees[req.Method]; !ok {
		r.HandleNotFound(w, req, r.middleware)
		return
	}

	nodes := r.trees[req.Method].Search(requestURL, false)
	if len(nodes) > 0 {
		node := nodes[0]

		if node.handle != nil {
			if node.path == requestURL {
				handle(w, req, node.handle, node.middleware)
				return
			}
			if node.path == requestURL[1:] {
				handle(w, req, node.handle, node.middleware)
			}
		}
	}

	if len(nodes) == 0 {
		list := strings.Split(requestURL, "/")
		prefix := list[1]
		nodes := r.trees[req.Method].Search(prefix, true)
		for _, node := range nodes {
			if handler := node.handle; handler != nil && node.path != requestURL {
				if matchParamsMap, ok := r.matchAndParse(requestURL, node.path); ok {
					ctx := context.WithValue(req.Context(), contextKey, matchParamsMap)
					req = req.WithContext(ctx)
					handle(w, req, handler, node.middleware)
					return
				}
			}
		}
	}

	r.HandleNotFound(w, req, r.middleware)
}

// HandleNotFound registers a handler when the request route is not found.
func (r *Router) HandleNotFound(w http.ResponseWriter, req *http.Request, middleware []Middleware) {
	if r.notFound != nil {
		handle(w, req, r.notFound, middleware)
		return
	}

	http.NotFound(w, req)
}

// handle executes middleware chain.
func handle(w http.ResponseWriter, req *http.Request, handler http.HandlerFunc, middleware []Middleware) {
	var baseHandler = handler
	for _, m := range middleware {
		baseHandler = m(baseHandler)
	}

	baseHandler(w, req)
}

// Match checks if the request matches the route pattern.
func (r *Router) Match(requestURL string, path string) bool {
	_, ok := r.matchAndParse(requestURL, path)
	return ok
}

// matchAndParse checks if the request matches the route and returns a map of the parse ones,
func (r *Router) matchAndParse(requestURL string, path string) (matchParams paramsMapType, b bool) {
	var (
		matchName []string
		pattern   string
	)

	b = true
	matchParams = make(paramsMapType)

	list := strings.Split(path, "/")
	for _, str := range list {
		if str == "" {
			continue
		}

		strLen := len(str)
		firstChar := str[0]
		lastChar := str[strLen-1]
		if string(firstChar) == "{" && string(lastChar) == "}" {
			matchStr := string(str[1 : strLen-1])
			list := strings.Split(matchStr, ":")
			pattern = pattern + "/" + "(" + list[1] + ")"
		} else if string(firstChar) == ":" {
			matchStr := str
			list := strings.Split(matchStr, ":")
			matchName = append(matchName, list[1])
			if list[1] == idKey {
				pattern = pattern + "/" + "(" + idPattern + ")"
			} else {
				pattern = pattern + "/" + "(" + defaultPattern + ")"
			}
		} else {
			pattern = pattern + "/" + str
		}
	}

	if strings.HasSuffix(requestURL, "/") {
		pattern = pattern + "/"
	}

	regex := regexp.MustCompile(pattern)
	if subMatch := regex.FindSubmatch([]byte(requestURL)); subMatch != nil {
		if string(subMatch[0]) == requestURL {
			subMatch = subMatch[:1]
			for k, v := range subMatch {
				matchParams[matchName[k]] = string(v)
			}
			return
		}
	}

	return nil, false
}
