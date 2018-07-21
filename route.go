package dgrouter

// Route is a command route
type Route struct {
	// Routes is a slice of subroutes
	Routes []*Route

	Name        string
	Description string
	Category    string

	// Matcher is a function that determines
	// If this route will be matched
	Matcher func(string) bool

	// Handler is the Handler for this route
	Handler HandlerFunc

	// Default route for responding to bot mentions
	Default *Route
}

// Desc sets this routes description
func (r *Route) Desc(description string) *Route {
	r.Description = description
	return r
}

// Cat sets this route's category
func (r *Route) Cat(category string) *Route {
	r.Category = category
	return r
}
