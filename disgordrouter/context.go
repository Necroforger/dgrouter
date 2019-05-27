package disgordrouter

import (
	"fmt"
	"sync"

	"github.com/andersfylling/snowflake/v3"

	"github.com/andersfylling/disgord"

	"github.com/Necroforger/dgrouter"
)

// Context represents a command context
type Context struct {
	// Route is the route that this command came from
	Route *dgrouter.Route
	Msg   *disgord.Message
	Ses   disgord.Session

	// List of arguments supplied with the command
	Args Args

	// Vars that can be optionally set using the Set and Get functions
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
func (c *Context) Reply(args ...interface{}) (*disgord.Message, error) {
	return c.Ses.SendMsg(c.Msg.ChannelID, fmt.Sprint(args...))
}

// ReplyEmbed replies to the sender with an embed
func (c *Context) ReplyEmbed(args ...interface{}) (*disgord.Message, error) {
	return c.Ses.CreateMessage(c.Msg.ChannelID, &disgord.CreateMessageParams{
		Embed: &disgord.Embed{
			Description: fmt.Sprint(args...),
		},
	})
}

// Guild retrieves a guild from the state or restapi
func (c *Context) Guild(guildID string) (*disgord.Guild, error) {
	return c.Ses.GetGuild(snowflake.ParseSnowflakeString(guildID))
}

// Channel retrieves a channel from the state or restapi
func (c *Context) Channel(channelID string) (*disgord.Channel, error) {
	return c.Ses.GetChannel(snowflake.ParseSnowflakeString(channelID))
}

// Member retrieves a member from the state or restapi
func (c *Context) Member(guildID, userID string) (*disgord.Member, error) {
	return c.Ses.GetMember(snowflake.ParseSnowflakeString(guildID), snowflake.ParseSnowflakeString(userID))
}

// NewContext returns a new context from a message
func NewContext(s disgord.Session, m *disgord.Message, args Args, route *dgrouter.Route) *Context {
	return &Context{
		Route: route,
		Msg:   m,
		Ses:   s,
		Args:  args,
		Vars:  map[string]interface{}{},
	}
}
