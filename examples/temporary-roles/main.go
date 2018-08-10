package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

var (
	fToken  = flag.String("t", "", "token to log into bot with")
	fDB     = flag.String("db", "roles.db", "location of boltdb database file to store data in")
	fPrefix = flag.String("p", "!", "bot prefix")
)

var database *bolt.DB

const bucket = "roles"

func monitorDatabase(s *discordgo.Session, done chan int) {
	t := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-t.C:
			database.Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists([]byte(bucket))
				if err != nil {
					return err
				}

				b.ForEach(func(k, v []byte) error {
					var r RoleExpiration
					err := json.Unmarshal(v, &r)
					if err != nil {
						log.Println("Error unmarshaling JSON: ", err)
						return err
					}
					if time.Now().After(r.Expires) {
						err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, r.RoleID)
						if err != nil {
							log.Println("error removing role")
							return err
						}
						err = b.Delete(k)
						if err != nil {
							log.Println("error deleting: ", err)
							return err
						}
					}
					return nil
				})

				return nil
			})
		case <-done:
			return
		}
	}
}

func main() {
	flag.Parse()

	s, err := discordgo.New(*fToken)
	if err != nil {
		log.Fatal(err)
	}

	database, err = bolt.Open(*fDB, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan int) // close this channel when you are done monitoring
	go monitorDatabase(s, done)

	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}

	r := exrouter.New()
	r.On("setrole", cmdRole).Desc("usage: setrole [userid] [role_name] [duration in seconds]")

	// Create help route and set it to the default route for bot mentions
	r.Default = r.On("help", func(ctx *exrouter.Context) {
		var text = ""
		var maxlen int
		for _, v := range r.Routes {
			if len(v.Name) > maxlen {
				maxlen = len(v.Name)
			}
		}
		for _, v := range r.Routes {
			text += fmt.Sprintf("%-"+strconv.Itoa(maxlen+5)+"s:    %s\n", v.Name, v.Description)
		}
		ctx.Reply("```" + text + "```")
	}).Desc("prints this help menu")

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		r.FindAndExecute(s, *fPrefix, s.State.User.ID, m.Message)
	})

	log.Println("bot is running...")
	// Prevent the bot from exiting
	<-make(chan struct{})
}

// RoleExpiration ...
type RoleExpiration struct {
	RoleID  string
	UserID  string
	GuildID string
	Expires time.Time
}

func saveRoleExpiration(data RoleExpiration) error {
	return database.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(id))

		s, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return b.Put(key, s)
	})
}

func cmdRole(ctx *exrouter.Context) {
	if ctx.Args.Get(1) == "" || ctx.Args.Get(2) == "" {
		ctx.Reply(ctx.Route.Description)
		return
	}

	guild, err := ctx.Guild(ctx.Msg.GuildID)
	if err != nil {
		ctx.Reply("Guild not found")
		return
	}

	role := findRoleByName(guild, ctx.Args.Get(2))
	if role == nil {
		ctx.Reply("Role not found")
		return
	}

	err = ctx.Ses.GuildMemberRoleAdd(ctx.Msg.GuildID, ctx.Args.Get(1), role.ID)
	if err != nil {
		ctx.Reply("Could not add role to member: ", err)
		return
	}

	var duration int
	if n, err := strconv.Atoi(ctx.Args.Get(3)); err == nil {
		duration = n
	} else {
		duration = 10
	}

	expires := time.Now().Add(time.Second * time.Duration(duration))

	saveRoleExpiration(RoleExpiration{
		Expires: expires,
		GuildID: ctx.Msg.GuildID,
		UserID:  ctx.Args.Get(1),
		RoleID:  role.ID,
	})

	ctx.Reply("Set temporary role for user\nIt will expire on " + expires.String())
}

func findRoleByName(guild *discordgo.Guild, name string) *discordgo.Role {
	for _, v := range guild.Roles {
		if strings.ToLower(v.Name) == name {
			return v
		}
	}
	return nil
}
