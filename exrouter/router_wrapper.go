package exrouter

import (
	"strings"

	"github.com/Necroforger/dgrouter"
	"github.com/bwmarrin/discordgo"
)

// HandlerFunc ...
type HandlerFunc func(*Context)

// Route wraps dgrouter.Router to use a Context
type Route struct {
	*dgrouter.Route
}

// New returns a new router wrapper
func New() *Route {
	return &Route{
		Route: dgrouter.New(),
	}
}

// On registers a handler function
func (r *Route) On(name string, handler HandlerFunc) *Route {
	return &Route{r.Route.On(name, WrapHandler(handler))}
}

// OnReg adds a route with a regular expression
func (r *Route) OnReg(name string, reg string, handler HandlerFunc) *Route {
	return &Route{r.Route.OnReg(name, reg, WrapHandler(handler))}
}

func mention(id string) string {
	return "<@" + id + ">"
}

func nickMention(id string) string {
	return "<@!" + id + ">"
}

// FindAndExecute is a helper method for calling routes
// it creates a context from a message, finds its route, and executes the handler
// it looks for a message prefix which is either the prefix specified or the message is prefixed
// with a bot mention
//    s            : discordgo session to pass to context
//    prefix       : prefix you want the bot to respond to
//    botID        : user ID of the bot to allow you to substitute the bot ID for a prefix
//    m            : discord message to pass to context
//    mentionRoute : route to serve when the bot recieves nothing but a mention
func (r *Route) FindAndExecute(s *discordgo.Session, prefix string, botID string, m *discordgo.Message) error {
	var pf string

	// If the message content is only a bot mention and the mention route is not nil, send the mention route
	if r.Default != nil && m.Content == mention(botID) || m.Content == nickMention(botID) {
		r.Default.Handler(NewContext(s, m, []string{""}, r.Default))
		return nil
	}

	// Append a space to the mentions
	bmention := mention(botID) + " "
	nmention := nickMention(botID) + " "

	p := func(t string) bool {
		return strings.HasPrefix(m.Content, t)
	}

	switch {
	case p(prefix):
		pf = prefix
	case p(bmention):
		pf = bmention
	case p(nmention):
		pf = nmention
	default:
		return dgrouter.ErrCouldNotFindRoute
	}

	command := strings.TrimPrefix(m.Content, pf)
	args := ParseArgs(command)

	if rt := r.Find(args.Get(0)); rt != nil {

		// Search through subroutes
		if len(args) > 1 {
			var i = 1
			var r = rt
			for _, v := range args[1:] {
				nr := r.Find(v)
				if nr == nil {
					break
				}
				r = nr
				i++
			}
			if r != nil {
				// Merge all arguments up to the subroute
				args = append(
					[]string{
						strings.Join(args[:i], string(separator)),
					},
					args[i:]...,
				)
				rt = r
			}
		}
		if rt.Handler != nil {
			ctx := NewContext(s, m, args, rt)
			rt.Handler(ctx)
		}
	} else {
		return dgrouter.ErrCouldNotFindRoute
	}

	return nil
}

// WrapHandler wraps a dgrouter.HandlerFunc
func WrapHandler(fn HandlerFunc) dgrouter.HandlerFunc {
	if fn == nil {
		return nil
	}
	return func(i interface{}) {
		fn(i.(*Context))
	}
}
