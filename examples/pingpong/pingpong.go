package main

import (
	"flag"
	"log"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

// Command line flags
var (
	fToken  = flag.String("t", "", "bot token")
	fPrefix = flag.String("p", "!", "bot prefix")
)

func main() {
	flag.Parse()

	s, err := discordgo.New(*fToken)
	if err != nil {
		log.Fatal(err)
	}

	router := exrouter.New()

	// Add some commands
	router.On("ping", func(ctx *exrouter.Context) {
		ctx.Reply("pong")
	}).Desc("responds with pong")

	router.On("avatar", func(ctx *exrouter.Context) {
		ctx.Reply(ctx.Msg.Author.AvatarURL("2048"))
	}).Desc("returns the user's avatar")

	helpRoute := router.On("help", func(ctx *exrouter.Context) {
		var text = ""
		for _, v := range router.Routes {
			text += v.Name + " : \t" + v.Description + "\n"
		}
		ctx.Reply("```" + text + "```")
	}).Desc("prints this help menu")

	// Add message handler
	s.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(s, *fPrefix, s.State.User.ID, m.Message, helpRoute)
	})

	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bot is running...")
	// Prevent the bot from exiting
	<-make(chan struct{})
}
