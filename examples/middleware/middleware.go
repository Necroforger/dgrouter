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

// AllowedNames are names allowed to use the auth commands
var AllowedNames = []string{
	"necroforger",
	"necro",

	"foxtail-grass-studios",

	"wriggle",
	"reimu",
	"marisa",
	"remilia",
	"flandre",
	"satori",
	"koishi",
	"parsee",
	"cirno",
}

// Auth is an authentication middleware
// Only allowing people with certain names to use these
// Routes
func Auth(fn exrouter.HandlerFunc) exrouter.HandlerFunc {
	return func(ctx *exrouter.Context) {
		member, err := ctx.Member(ctx.Msg.GuildID, ctx.Msg.Author.ID)
		if err != nil {
			ctx.Reply("Could not fetch member: ", err)
		}

		ctx.Reply("Authenticating...")

		for _, v := range AllowedNames {
			if member.Nick == v {
				ctx.Set("member", member)
				fn(ctx)
				return
			}
		}

		ctx.Reply("You don't have permission to use this command")
	}
}

func main() {
	flag.Parse()

	s, err := discordgo.New(*fToken)
	if err != nil {
		log.Fatal(err)
	}

	router := exrouter.New()

	router.Group(func(r *exrouter.Route) {
		// Added routes inherit their parent category.
		// I set the parent category here and it won't affect the
		// Actual router, just this group
		r.Cat("main")

		// This authentication middleware applies only to this group
		r.Use(Auth)
		log.Printf("len(middleware) = %d\n", len(r.Middleware))

		r.On("testauth", func(ctx *exrouter.Context) {
			ctx.Reply("Hello " + ctx.Get("member").(*discordgo.Member).Nick + ", you have permission to use this command")
		})
	})

	router.Group(func(r *exrouter.Route) {
		r.Cat("other")
		r.On("ping", func(ctx *exrouter.Context) { ctx.Reply("pong") }).Desc("Responds with pong").Cat("other")
	})

	router.Default = router.On("help", func(ctx *exrouter.Context) {
		var text = ""
		for _, v := range router.Routes {
			text += v.Name + " : \t" + v.Description + ":\t" + v.Category + "\n"
		}
		ctx.Reply("```" + text + "```")
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
