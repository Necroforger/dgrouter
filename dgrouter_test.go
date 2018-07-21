package dgrouter_test

import (
	"log"
	"testing"

	"github.com/Necroforger/dgrouter"
)

func TestRouter(t *testing.T) {
	r := dgrouter.New()

	r.On("ping", func(i interface{}) { log.Println("hello") }).Desc("Responds with pong").Cat("general")
	r.OnReg("hello", "h.llo", nil).Desc("tests regular expressions").Cat("regex")

	if rt := r.Find("ping"); rt != nil {
		rt.Handler(nil)
	} else {
		t.Fail()
	}

	if rt := r.Find("route that doesn't exist"); rt != nil {
		t.Fail()
	}

	if rt := r.Find("hallo"); rt != nil {
		log.Println("found route")
	} else {
		t.Fail()
	}
}
