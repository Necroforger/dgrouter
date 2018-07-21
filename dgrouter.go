package dgrouter

import (
	"errors"
)

// Error variables
var (
	ErrCouldNotFindRoute  = errors.New("Could not find route")
	ErrRouteAlreadyExists = errors.New("route already exists")
)

// HandlerFunc is a command handler
type HandlerFunc func(interface{})

// On registers a route with the name you supply
//    name    : name of the route to create
//    handler : handler function
func (r *Route) On(name string, handler HandlerFunc) *Route {
	if rt := r.Find(name); rt != nil {
		return rt
	}

	route := &Route{}
	route.Name = name
	route.Matcher = NewNameMatcher(route)
	route.Handler = handler

	r.AddRoute(route)
	return route
}

// OnReg registers a route with the name and regular expression that you supply
//    name    : name of the route to create
//    regex   : regular expression to match
//    handler : handler function for the route
func (r *Route) OnReg(name, regex string, handler HandlerFunc) *Route {
	if rt := r.Find(name); rt != nil {
		return rt
	}

	route := &Route{}
	route.Name = name
	route.Matcher = NewRegexMatcher(regex)
	route.Handler = handler

	r.AddRoute(route)
	return route
}

// AddRoute adds a route to the router
// Will return RouteAlreadyExists error on failure
//    route : route to add
func (r *Route) AddRoute(route *Route) error {
	// Check if the route already exists
	if rt := r.Find(route.Name); rt != nil {
		return ErrRouteAlreadyExists
	}

	r.Routes = append(r.Routes, route)
	return nil
}

// RemoveRoute removes a route from the router
//     route : route to remove
func (r *Route) RemoveRoute(route *Route) error {
	for i, v := range r.Routes {
		if v == route {
			r.Routes = append(r.Routes[:i], r.Routes[i+1:]...)
			return nil
		}
	}
	return ErrCouldNotFindRoute
}

// Find finds a route with the given name
// It will return nil if nothing is found
//    name : name of route to find
func (r *Route) Find(name string) *Route {
	for _, v := range r.Routes {
		if v.Matcher(name) {
			return v
		}
	}
	return nil
}

// FindFull a full path of routes by searching through their subroutes
// Until the deepest match is found.
// It will return the route matched and the depth it was found at
//     args : path of route you wish to find
//            ex. FindFull(command, subroute1, subroute2, nonexistent)
//            will return the deepest found match, which will be subroute2
func (r *Route) FindFull(args ...string) (*Route, int) {
	nr := r
	i := 0
	for _, v := range args {
		if rt := nr.Find(v); rt != nil {
			nr = rt
			i++
		} else {
			break
		}
	}
	return nr, i
}

// New returns a new route
func New() *Route {
	return &Route{
		Routes: []*Route{},
	}
}
