package exmiddleware

import (
	"errors"
	"time"

	"github.com/Necroforger/dgrouter"
	"github.com/Necroforger/dgrouter/exrouter"
)

const (
	ctxPrefix  = "middleware"
	ctxError   = ctxPrefix + ".err"
	ctxGuild   = ctxPrefix + ".guild"
	ctxChannel = ctxPrefix + ".channel"
)

// Errors
var (
	ErrOnCooldown     = errors.New("command is on cooldown")
	ErrChannelNotNSFW = errors.New("this command can only be used in an NSFW channel")
)

// CatchFunc function called if one of the middleware experiences an error
// Can be left as nil
type CatchFunc func(ctx *exrouter.Context)

// DefaultCatch is the default catch function
func DefaultCatch() func(ctx *exrouter.Context) {
	return func(ctx *exrouter.Context) {
		if e := Err(ctx); e != nil {
			ctx.Reply("error: ", e)
		}
	}
}

// CatchReply returns a function that prints the message you pass it
func CatchReply(message string) func(ctx *exrouter.Context) {
	return func(ctx *exrouter.Context) {
		ctx.Reply(message)
	}
}

// UserCooldown creates a user specific cooldown timer for a route or collection of routes
func UserCooldown(cooldown time.Duration, catch CatchFunc) exrouter.MiddlewareFunc {
	// Table is a map of userIDs to a map of routes which store the last time they were called
	// By a given user.
	table := map[string]map[*dgrouter.Route]time.Time{}
	return func(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
		return func(ctx *exrouter.Context) {
			user, ok := table[ctx.Msg.Author.ID]
			if !ok {
				table[ctx.Msg.Author.ID] = map[*dgrouter.Route]time.Time{}
				return
			}

			// Retrieve the last time this command was used
			last, ok := user[ctx.Route]
			if !ok {
				// Set the last time the command was used and return
				// If nothing was found
				user[ctx.Route] = time.Now()
				fn(ctx)
				return
			}

			if !time.Now().After(last.Add(cooldown)) {
				callCatch(ctx, catch, ErrOnCooldown)
				return
			}

			// Update the last time command was used
			user[ctx.Route] = time.Now()
			fn(ctx)
		}
	}
}

// RequireNSFW requires a message to be sent from an NSFW channel
func RequireNSFW(catch CatchFunc) exrouter.MiddlewareFunc {
	return func(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
		return func(ctx *exrouter.Context) {
			channel, err := getChannel(ctx.Ses, ctx.Msg.ChannelID)
			if err != nil {
				callCatch(ctx, catch, err)
				return
			}
			if !channel.NSFW {
				callCatch(ctx, catch, ErrChannelNotNSFW)
				return
			}
			ctx.Set(ctxChannel, channel)
			fn(ctx)
		}
	}
}

// GetGuild retrieves the guild in which the message was sent from
func GetGuild(catch CatchFunc) exrouter.MiddlewareFunc {
	return func(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
		return func(ctx *exrouter.Context) {
			guild, err := getGuild(ctx.Ses, ctx.Msg.GuildID)
			if err != nil {
				callCatch(ctx, catch, err)
				return
			}

			ctx.Set(ctxGuild, guild)
			fn(ctx)
		}
	}
}

// GetChannel retrieves the channel in which the message was sent from
func GetChannel(catch CatchFunc) exrouter.MiddlewareFunc {
	return func(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
		return func(ctx *exrouter.Context) {
			channel, err := getChannel(ctx.Ses, ctx.Msg.GuildID)
			if err != nil {
				callCatch(ctx, catch, err)
				return
			}

			ctx.Set(ctxChannel, channel)
			fn(ctx)
		}
	}
}
