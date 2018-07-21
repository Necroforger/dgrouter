package exrouter

import (
	"fmt"
	"sync"

	"github.com/Necroforger/dgrouter"

	"github.com/bwmarrin/discordgo"
)

// Context represents a command context
type Context struct {
	Route *dgrouter.Route
	Msg   *discordgo.Message
	Ses   *discordgo.Session
	Args  Args

	vmu  sync.RWMutex
	Vars map[string]interface{}
}

// Set sets a variable on the context
func (c *Context) Set(key string, d interface{}) {
	c.vmu.Lock()
	c.Vars[key] = d
	c.vmu.Unlock()
}

// Get retrieves a variable from the context
func (c *Context) Get(key string) interface{} {
	if c, ok := c.Vars[key]; ok {
		return c
	}
	return nil
}

// Reply replies to the sender with the given message
func (c *Context) Reply(args ...interface{}) (*discordgo.Message, error) {
	return c.Ses.ChannelMessageSend(c.Msg.ChannelID, fmt.Sprint(args...))
}

// ReplyEmbed replies to the sender with an embed
func (c *Context) ReplyEmbed(args ...interface{}) (*discordgo.Message, error) {
	return c.Ses.ChannelMessageSendEmbed(c.Msg.ChannelID, &discordgo.MessageEmbed{
		Description: fmt.Sprint(args...),
	})
}

// NewContext returns a new context from a message
func NewContext(s *discordgo.Session, m *discordgo.Message, args Args, route *dgrouter.Route) *Context {
	return &Context{
		Route: route,
		Msg:   m,
		Ses:   s,
		Args:  args,
		Vars:  map[string]interface{}{},
	}
}
