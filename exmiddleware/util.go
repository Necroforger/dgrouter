package exmiddleware

import (
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

func getChannel(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	channel, err := s.State.Channel(channelID)
	if err != nil {
		return s.Channel(channelID)
	}
	return channel, err
}

func getGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return s.Guild(guildID)
	}
	return guild, err
}

func getMember(s *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		return s.GuildMember(guildID, userID)
	}
	return member, err
}

// callCatch calls a catch function with an error
func callCatch(ctx *exrouter.Context, fn CatchFunc, err error) {
	if fn == nil {
		return
	}
	ctx.Set(ctxError, err)
	fn(ctx)
}
