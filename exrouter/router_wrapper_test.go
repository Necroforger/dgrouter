package exrouter_test

import (
	"log"
	"testing"

	"github.com/Necroforger/dgrouter"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

func TestRouter(t *testing.T) {
	messages := []string{
		"!ping",
		"!say hello",
		"!test args one two three",
		"<@botid> say hello",
		"<@!botid> say hello",
		"<@!botid>",
	}

	r := exrouter.Route{
		Route: dgrouter.New(),
	}

	r.On("ping", func(ctx *exrouter.Context) {})

	r.On("say", func(ctx *exrouter.Context) {
		if ctx.Args.Get(1) != "hello" {
			t.Fail()
		}
	})

	r.On("test", func(ctx *exrouter.Context) {
		ctx.Set("hello", "hi")
		if r := ctx.Get("hello"); r.(string) != "hi" {
			t.Fail()
		}
		expected := []string{"args", "one", "two", "three"}
		for i, v := range expected {
			if ctx.Args.Get(i+1) != v {
				t.Fail()
			}
		}
	})

	mentionRoute := r.On("help", func(ctx *exrouter.Context) {
		log.Println("Bot was mentioned")
	})

	// Set the default route for this router
	// Will be triggered on bot mentions
	r.Handler = mentionRoute.Handler

	for _, v := range messages {
		// Construct mock message
		msg := &discordgo.Message{
			Author: &discordgo.User{
				Username: "necroforger",
				Bot:      false,
			},
			Content: v,
		}

		// Attempt to find and execute the route for this message
		err := r.FindAndExecute(nil, "!", "botid", msg)
		if err != nil {
			t.Fail()
		}
	}
}
