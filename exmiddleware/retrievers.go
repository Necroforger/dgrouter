package exmiddleware

import (
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

// Err retrieves the error variable from the context
func Err(ctx *exrouter.Context) error {
	if v := ctx.Get(ctxError); v != nil {
		return v.(error)
	}
	return nil
}

// Guild retrieves the guild variable from a context
func Guild(ctx *exrouter.Context) *discordgo.Guild {
	if v := ctx.Get(ctxGuild); v != nil {
		return v.(*discordgo.Guild)
	}
	return nil
}

// Channel retrieves the channel variable from a context
func Channel(ctx *exrouter.Context) *discordgo.Channel {
	if v := ctx.Get(ctxChannel); v != nil {
		return v.(*discordgo.Channel)
	}
	return nil
}

// Member fetches the member from the context
func Member(ctx *exrouter.Context) *discordgo.Member {
	if v := ctx.Get(ctxMember); v != nil {
		return v.(*discordgo.Member)
	}
	return nil
}
