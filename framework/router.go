package framework

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Handler processes one routed update.
type Handler func(*Context) error

// Middleware wraps a handler. Middleware declared first is outermost.
type Middleware func(Handler) Handler

type routeDef struct {
	filter     Filter
	handler    Handler
	middleware []Middleware
}

type compiledRoute struct {
	filter  Filter
	handler Handler
}

type prefixRouteDef struct {
	prefix string
	route  routeDef
}

type compiledPrefixTable struct {
	routes  map[string][]compiledRoute
	lengths []int
}

type routeTable struct {
	rawCommands         map[string][]routeDef
	rawCallbacks        map[string][]routeDef
	rawCallbackPrefixes []prefixRouteDef
	rawFilters          []routeDef
	rawFallback         Handler
	middleware          []Middleware

	commands         map[string][]compiledRoute
	callbacks        map[string][]compiledRoute
	callbackPrefixes *compiledPrefixTable
	filters          []compiledRoute
	fallback         Handler
}

// Router performs concurrency-safe command, callback, prefix, and filter
// dispatch. Reads use immutable snapshots; registration remains safe after
// dispatch starts.
type Router struct {
	mu      sync.Mutex
	table   atomic.Pointer[routeTable]
	startup *routeTable
	started atomic.Bool
}

// NewRouter creates an empty router.
func NewRouter() *Router {
	r := &Router{}
	initial := &routeTable{
		rawCommands:  make(map[string][]routeDef),
		rawCallbacks: make(map[string][]routeDef),
		commands:     make(map[string][]compiledRoute),
		callbacks:    make(map[string][]compiledRoute),
	}
	r.table.Store(initial)
	r.startup = cloneRouteTable(initial)
	return r
}

// Command registers an exact slash command. A later direct registration for
// the same command replaces the earlier direct handler.
func (r *Router) Command(command string, handler Handler) {
	command = normalizeCommand(command)
	if command == "" || handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawCommands[command] = []routeDef{{handler: handler}}
	})
}

// Callback registers an exact callback-data handler.
func (r *Router) Callback(data string, handler Handler) {
	if handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawCallbacks[data] = []routeDef{{handler: handler}}
	})
}

// CallbackPrefix registers a callback-data prefix. The longest matching
// prefix wins; equal-length prefixes preserve registration order.
func (r *Router) CallbackPrefix(prefix string, handler Handler) {
	if prefix == "" || handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		filtered := table.rawCallbackPrefixes[:0]
		for _, item := range table.rawCallbackPrefixes {
			if item.prefix != prefix {
				filtered = append(filtered, item)
			}
		}
		table.rawCallbackPrefixes = append(filtered, prefixRouteDef{
			prefix: prefix,
			route:  routeDef{handler: handler},
		})
	})
}

// On registers an ordered filtered route. The first matching route handles the update.
func (r *Router) On(filter Filter, handler Handler) {
	if handler == nil {
		return
	}
	r.addRoute(routeDef{filter: filter, handler: handler})
}

// Use appends global middleware in declaration order.
func (r *Router) Use(middleware ...Middleware) {
	r.mutate(func(table *routeTable) {
		for _, item := range middleware {
			if item != nil {
				table.middleware = append(table.middleware, item)
			}
		}
	})
}

// OnUpdate sets the fallback handler used when no route matches.
func (r *Router) OnUpdate(handler Handler) {
	r.mutate(func(table *routeTable) { table.rawFallback = handler })
}

// Group creates a route group with shared filters.
func (r *Router) Group(filters ...Filter) *Group {
	return &Group{router: r, filter: All(filters...)}
}

// Handle routes one context synchronously. At most one handler runs.
func (r *Router) Handle(c *Context) error {
	if r == nil || c == nil {
		return nil
	}
	r.ensureStarted()
	table := r.table.Load()
	if table == nil {
		return nil
	}

	if command := c.Command(); command != "" {
		if handler := matchRoutes(table.commands[command], c); handler != nil {
			return handler(c)
		}
	}

	if c.Callback != nil {
		if handler := matchRoutes(table.callbacks[c.Callback.Data], c); handler != nil {
			return handler(c)
		}
		if handler := matchCallbackPrefix(table.callbackPrefixes, c.Callback.Data, c); handler != nil {
			return handler(c)
		}
	}

	if handler := matchRoutes(table.filters, c); handler != nil {
		return handler(c)
	}
	if table.fallback != nil {
		return table.fallback(c)
	}
	return nil
}

func (r *Router) addCommand(command string, route routeDef) {
	command = normalizeCommand(command)
	if command == "" || route.handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawCommands[command] = append(table.rawCommands[command], route)
	})
}

func (r *Router) addCallback(data string, route routeDef) {
	if route.handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawCallbacks[data] = append(table.rawCallbacks[data], route)
	})
}

func (r *Router) addCallbackPrefix(prefix string, route routeDef) {
	if prefix == "" || route.handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawCallbackPrefixes = append(table.rawCallbackPrefixes, prefixRouteDef{
			prefix: prefix,
			route:  route,
		})
	})
}

func (r *Router) addRoute(route routeDef) {
	if route.handler == nil {
		return
	}
	r.mutate(func(table *routeTable) {
		table.rawFilters = append(table.rawFilters, route)
	})
}

func (r *Router) mutate(change func(*routeTable)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started.Load() {
		if r.startup == nil {
			r.startup = cloneRouteTable(r.table.Load())
		}
		change(r.startup)
		return
	}

	next := cloneRouteTable(r.table.Load())
	change(next)
	next.rebuild()
	r.table.Store(next)
}

func (r *Router) ensureStarted() {
	if r == nil || r.started.Load() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started.Load() {
		return
	}
	next := r.startup
	if next == nil {
		next = cloneRouteTable(r.table.Load())
	}
	next.rebuild()
	r.table.Store(next)
	r.startup = nil
	r.started.Store(true)
}

func (t *routeTable) rebuild() {
	t.commands = compileRouteMap(t.rawCommands, t.middleware)
	t.callbacks = compileRouteMap(t.rawCallbacks, t.middleware)
	t.filters = compileRoutes(t.rawFilters, t.middleware)
	t.callbackPrefixes = compileCallbackPrefixes(t.rawCallbackPrefixes, t.middleware)
	t.fallback = wrap(t.rawFallback, t.middleware)
}

func compileCallbackPrefixes(routes []prefixRouteDef, global []Middleware) *compiledPrefixTable {
	if len(routes) == 0 {
		return nil
	}
	table := &compiledPrefixTable{routes: make(map[string][]compiledRoute, len(routes))}
	lengths := make(map[int]struct{})
	for _, route := range routes {
		table.routes[route.prefix] = append(
			table.routes[route.prefix],
			compileRoute(route.route, global),
		)
		if _, exists := lengths[len(route.prefix)]; !exists {
			lengths[len(route.prefix)] = struct{}{}
			table.lengths = append(table.lengths, len(route.prefix))
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(table.lengths)))
	return table
}

func compileRouteMap(source map[string][]routeDef, global []Middleware) map[string][]compiledRoute {
	target := make(map[string][]compiledRoute, len(source))
	for key, routes := range source {
		target[key] = compileRoutes(routes, global)
	}
	return target
}

func compileRoutes(routes []routeDef, global []Middleware) []compiledRoute {
	result := make([]compiledRoute, 0, len(routes))
	for _, route := range routes {
		if route.handler != nil {
			result = append(result, compileRoute(route, global))
		}
	}
	return result
}

func compileRoute(route routeDef, global []Middleware) compiledRoute {
	chain := make([]Middleware, 0, len(global)+len(route.middleware))
	chain = append(chain, global...)
	chain = append(chain, route.middleware...)
	return compiledRoute{filter: route.filter, handler: wrap(route.handler, chain)}
}

func cloneRouteTable(source *routeTable) *routeTable {
	if source == nil {
		return &routeTable{rawCommands: make(map[string][]routeDef), rawCallbacks: make(map[string][]routeDef)}
	}
	target := &routeTable{
		rawCommands:         cloneRouteMap(source.rawCommands),
		rawCallbacks:        cloneRouteMap(source.rawCallbacks),
		rawCallbackPrefixes: append([]prefixRouteDef(nil), source.rawCallbackPrefixes...),
		rawFilters:          append([]routeDef(nil), source.rawFilters...),
		rawFallback:         source.rawFallback,
		middleware:          append([]Middleware(nil), source.middleware...),
	}
	return target
}

func cloneRouteMap(source map[string][]routeDef) map[string][]routeDef {
	target := make(map[string][]routeDef, len(source))
	for key, routes := range source {
		target[key] = append([]routeDef(nil), routes...)
	}
	return target
}

func matchRoutes(routes []compiledRoute, c *Context) Handler {
	for _, route := range routes {
		if routeMatches(route, c) {
			return route.handler
		}
	}
	return nil
}

func matchCallbackPrefix(table *compiledPrefixTable, data string, c *Context) Handler {
	if table == nil {
		return nil
	}
	for _, length := range table.lengths {
		if length > len(data) {
			continue
		}
		if handler := matchRoutes(table.routes[data[:length]], c); handler != nil {
			return handler
		}
	}
	return nil
}

func routeMatches(route compiledRoute, c *Context) bool {
	return route.handler != nil && (route.filter == nil || route.filter(c))
}

func wrap(handler Handler, middleware []Middleware) Handler {
	if handler == nil {
		return nil
	}
	result := handler
	for index := len(middleware) - 1; index >= 0; index-- {
		if middleware[index] != nil {
			result = middleware[index](result)
		}
	}
	return result
}

func normalizeCommand(command string) string {
	command = strings.TrimSpace(strings.TrimPrefix(command, "/"))
	command, _, _ = strings.Cut(command, "@")
	return strings.ToLower(command)
}

// Group applies shared filters and middleware to a set of routes without
// adding another dispatch layer.
type Group struct {
	mu         sync.RWMutex
	router     *Router
	filter     Filter
	middleware []Middleware
}

// Use appends group middleware and returns the group for chaining.
func (g *Group) Use(middleware ...Middleware) *Group {
	if g == nil {
		return g
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, item := range middleware {
		if item != nil {
			g.middleware = append(g.middleware, item)
		}
	}
	return g
}

// Group creates a nested group inheriting current filters and middleware.
func (g *Group) Group(filters ...Filter) *Group {
	if g == nil {
		return nil
	}
	filter, middleware := g.snapshot()
	return &Group{
		router:     g.router,
		filter:     All(filter, All(filters...)),
		middleware: middleware,
	}
}

// Command registers a command inside the group.
func (g *Group) Command(command string, handler Handler) {
	if g != nil && g.router != nil {
		g.router.addCommand(command, g.route(handler))
	}
}

// Callback registers an exact callback inside the group.
func (g *Group) Callback(data string, handler Handler) {
	if g != nil && g.router != nil {
		g.router.addCallback(data, g.route(handler))
	}
}

// CallbackPrefix registers a callback prefix inside the group.
func (g *Group) CallbackPrefix(prefix string, handler Handler) {
	if g != nil && g.router != nil {
		g.router.addCallbackPrefix(prefix, g.route(handler))
	}
}

// On registers an ordered filtered route inside the group.
func (g *Group) On(filter Filter, handler Handler) {
	if g != nil && g.router != nil {
		groupFilter, middleware := g.snapshot()
		g.router.addRoute(routeDef{
			filter:     All(groupFilter, filter),
			handler:    handler,
			middleware: middleware,
		})
	}
}

func (g *Group) route(handler Handler) routeDef {
	filter, middleware := g.snapshot()
	return routeDef{filter: filter, handler: handler, middleware: middleware}
}

func (g *Group) snapshot() (Filter, []Middleware) {
	if g == nil {
		return nil, nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.filter, append([]Middleware(nil), g.middleware...)
}
