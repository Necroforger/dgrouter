package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rylio/ytdl"

	"github.com/jonas747/dca"

	"sync"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

// Command line flags
var (
	fToken    = flag.String("t", "", "bot token")
	fPrefix   = flag.String("p", "!", "bot prefix")
	fSoundDir = flag.String("d", "sounds", "directory of sound files")
	fWatch    = flag.Bool("watch", false, "watch the sound directory for changes")
)

func reply(ctx *exrouter.Context, args ...interface{}) {
	_, err := ctx.Reply(args...)
	if err != nil {
		log.Println("error sending message: ", err)
	}
}

func decodeFromFile(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&v)
}

func trimExtension(p string) string {
	return strings.TrimSuffix(p, filepath.Ext(p))
}

func createRouter(s *discordgo.Session) *exrouter.Route {
	router := exrouter.New()

	// Create playback functions
	files, err := ioutil.ReadDir(*fSoundDir)
	if err != nil {
		log.Fatal("no sound directory: ", err)
	}
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		if filepath.Ext(v.Name()) == ".json" {
			var data = [][]string{}
			decodeFromFile(path.Join(*fSoundDir, v.Name()), &data)
			for _, d := range data {
				router.On(trimExtension(d[0]), createYoutubeFunction(d[1])).Desc("Plays " + d[1])
			}
			continue
		}
		router.On(
			trimExtension(v.Name()),
			createMusicFunction(filepath.Join(*fSoundDir, v.Name())),
		).Desc("plays " + v.Name())
	}

	router.On("stop", func(ctx *exrouter.Context) {
		stopStreaming(ctx.Msg.GuildID)
	}).Desc("Stops the currently running stream")

	router.On("leave", func(ctx *exrouter.Context) {
		s.Lock()
		if vc, ok := s.VoiceConnections[ctx.Msg.GuildID]; ok {
			err := vc.Disconnect()
			if err != nil {
				reply(ctx, "error disconnecting from voice channel: ", err)
			}
		}
		s.Unlock()
	}).Desc("Leaves the current voice channel")

	router.On("yt", func(ctx *exrouter.Context) {
		createYoutubeFunction(ctx.Args.Get(1))(ctx)
	}).Desc("plays a youtube link").Alias("youtube")

	// Create help route and set it to the default route for bot mentions
	router.Default = router.On("help", func(ctx *exrouter.Context) {
		var text = ""
		var maxlen int
		for _, v := range router.Routes {
			if len(v.Name) > maxlen {
				maxlen = len(v.Name)
			}
		}
		for _, v := range router.Routes {
			text += fmt.Sprintf("%-"+strconv.Itoa(maxlen+5)+"s:    %s\n", v.Name, v.Description)
		}
		reply(ctx, "```"+text+"```")
	}).Desc("prints this help menu")

	return router
}

func main() {
	flag.Parse()

	s, err := discordgo.New(*fToken)
	if err != nil {
		log.Fatal(err)
	}

	var rmu sync.RWMutex
	router := createRouter(s)

	// Add message handler
	s.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		rmu.RLock()
		defer rmu.RUnlock()
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

// thread safe map
type syncmap struct {
	sync.RWMutex
	Data map[string]interface{}
}

func newSyncmap() *syncmap {
	return &syncmap{
		Data: map[string]interface{}{},
	}
}

// Set ...
func (s *syncmap) Set(key string, value interface{}) {
	s.Lock()
	s.Data[key] = value
	s.Unlock()
}

// Get ...
func (s *syncmap) Get(key string) (interface{}, bool) {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.Data[key]; ok {
		return v, true
	}
	return nil, false
}

// links guild ids to their dca encoder session
var streams = newSyncmap()

type streamSession struct {
	EncodeSession *dca.EncodeSession
	StreamSession *dca.StreamingSession
}

func createMusicFunction(fpath string) func(ctx *exrouter.Context) {
	return func(ctx *exrouter.Context) {
		vc, err := getVoiceConnection(ctx.Ses, ctx.Msg.Author.ID, ctx.Msg.GuildID)
		if err != nil {
			reply(ctx, "error obtaining voice connection")
			log.Println("error getting voice connection: ", err)
			return
		}

		log.Println("Creating encoding session")

		opts := dca.StdEncodeOptions
		opts.RawOutput = true
		opts.Bitrate = 120

		encodeSession, err := dca.EncodeFile(fpath, opts)
		if err != nil {
			reply(ctx, "error creating encode session")
			log.Println("error creating encode session: ", err)
			return
		}

		// set speaking to true
		if err := vc.Speaking(true); err != nil {
			reply(ctx, "could not set speaking to true")
			return
		}
		defer vc.Speaking(false)

		done := make(chan error)
		streamer := dca.NewStream(encodeSession, vc, done)

		// Stop any currently running stream
		stopStreaming(vc.GuildID)

		log.Println("Adding streaming to map")
		streams.Set(ctx.Msg.GuildID, &streamSession{
			EncodeSession: encodeSession,
			StreamSession: streamer,
		})

		if err := <-done; err != nil {
			log.Println(err)
			// Clean up incase something happened and ffmpeg is still running
			encodeSession.Truncate()
		}
	}
}

// createYoutubeFunction creates a function for streaming youtube videos
func createYoutubeFunction(yurl string) func(ctx *exrouter.Context) {
	return func(ctx *exrouter.Context) {
		vc, err := getVoiceConnection(ctx.Ses, ctx.Msg.Author.ID, ctx.Msg.GuildID)
		if err != nil {
			reply(ctx, "error obtaining voice connection")
			log.Println("error getting voice connection: ", err)
			return
		}

		err = playYoutubeVideo(vc, yurl)
		if err != nil {
			reply(ctx, "error playing youtube video: ", err)
		}
	}
}

func playYoutubeVideo(vc *discordgo.VoiceConnection, yurl string) error {
	info, err := ytdl.GetVideoInfo(yurl)
	if err != nil {
		return err
	}

	rd, wr := io.Pipe()
	go func() {
		if key := info.Formats.Best(ytdl.FormatAudioEncodingKey); len(key) != 0 {
			err := info.Download(key[0], wr)
			if err != nil {
				log.Println(err)
			}
		}
		wr.Close()
	}()

	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 120

	encodeSession, err := dca.EncodeMem(rd, opts)
	if err != nil {
		return err
	}

	// set speaking to true
	if err := vc.Speaking(true); err != nil {
		return err
	}
	defer vc.Speaking(false)

	done := make(chan error)
	streamer := dca.NewStream(encodeSession, vc, done)

	// Stop any currently running stream
	stopStreaming(vc.GuildID)

	log.Println("Adding streaming to map")
	streams.Set(vc.GuildID, &streamSession{
		EncodeSession: encodeSession,
		StreamSession: streamer,
	})

	if err := <-done; err != nil {
		log.Println(err)
		// Clean up incase something happened and ffmpeg is still running
		encodeSession.Cleanup()
	}

	return nil
}

func stopStreaming(guildID string) {
	if v, ok := streams.Get(guildID); ok {
		// v.(*streamSession).EncodeSession.Cleanup()
		err := v.(*streamSession).EncodeSession.Stop()
		if err != nil {
			log.Println("error stopping streaming session: ", err)
		}
		v.(*streamSession).EncodeSession.Cleanup()
	}
}

// getVoiceConnection gets a bot's voice connection
func getVoiceConnection(s *discordgo.Session, userID string, guildID string) (*discordgo.VoiceConnection, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range guild.VoiceStates {
		if v.UserID == userID {
			return s.ChannelVoiceJoin(guildID, v.ChannelID, false, false)
		}
	}

	return nil, errors.New("Voice connection not found")
}
