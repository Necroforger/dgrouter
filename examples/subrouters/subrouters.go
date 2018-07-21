package main

import (
	"flag"
	"log"
	"strings"

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

	router.On("sub", nil).
		On("sub2", func(ctx *exrouter.Context) {
			ctx.Reply("sub2 called with arguments:\n", strings.Join(ctx.Args, ";"))
		}).
		On("sub3", func(ctx *exrouter.Context) {
			ctx.Reply("sub3 called with arguments:\n", strings.Join(ctx.Args, ";"))
		})

	router.Default = router.On("help", func(ctx *exrouter.Context) {
		var f func(depth int, r *exrouter.Route) string
		f = func(depth int, r *exrouter.Route) string {
			text := ""
			for _, v := range r.Routes {
				text += strings.Repeat("  ", depth) + v.Name + " : " + v.Description + "\n"
				text += f(depth+1, &exrouter.Route{Route: v})
			}
			return text
		}
		ctx.Reply("```" + f(0, router) + "```")
	}).Desc("prints this help menu")

	// Add message handler
	s.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(s, *fPrefix, s.State.User.ID, m.Message)
	})

	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bot is running...")
	// Prevent the bot from exiting
	<-make(chan struct{})
}
