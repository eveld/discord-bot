package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-hclog"
	wasp "github.com/nicholasjackson/wasp/engine"
	"github.com/nicholasjackson/wasp/engine/logger"
)

const Version = "v0.0.0-alpha"

const (
	MessageCreateFunc = "message_create"
)

var Token string
var Session *discordgo.Session
var pipeline *wasp.Wasm

var plugins []Plugin
var subscriptions map[discordgo.Intent][]Plugin

func init() {
	Token = os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		flag.StringVar(&Token, "t", "", "Discord Authentication Token")
		flag.Parse()
	}

	plugins = []Plugin{
		{
			Name:    "profanity",
			Version: "v0.0.1",
			Path:    "plugins/profanity/module.wasm",
			Events:  []discordgo.Intent{discordgo.IntentsGuildMessages},
		},
	}

	subscriptions = map[discordgo.Intent][]Plugin{}
}

func main() {
	log := hclog.Default()
	log.SetLevel(hclog.Info)

	var err error

	// ASCII art in the "Elite" font generated at https://patorjk.com/software/taag/
	fmt.Printf(`
·▄▄▄▄  ▪  .▄▄ ·  ▄▄·       ▄▄▄  ·▄▄▄▄  
██▪ ██ ██ ▐█ ▀. ▐█ ▌▪▪     ▀▄ █·██▪ ██ 
▐█· ▐█▌▐█·▄▀▀▀█▄██ ▄▄ ▄█▀▄ ▐▀▀▄ ▐█· ▐█▌
██. ██ ▐█▌▐█▄▪▐█▐███▌▐█▌.▐▌▐█•█▌██. ██ 
▀▀▀▀▀• ▀▀▀ ▀▀▀▀ ·▀▀▀  ▀█▄▀▪.▀  ▀▀▀▀▀▀• 
                          %-16s`+"\n\n", Version)

	Session, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	pipeline = wasp.New(logger.New(log.Info, log.Debug, log.Error, log.Trace))

	callbacks := &wasp.Callbacks{}
	callbacks.AddCallback("env", "send_channel_message", sendChannelMessage)
	callbacks.AddCallback("env", "delete_channel_message", deleteChannelMessage)

	pluginConfig := &wasp.PluginConfig{
		Callbacks: callbacks,
	}

	// Read configured plugins from configuration file
	// Loop over the plugins
	// Register each of the plugins
	for _, plugin := range plugins {
		err = pipeline.RegisterPlugin(plugin.Name, plugin.Path, pluginConfig)
		if err != nil {
			log.Error("Error loading plugin", "error", err)
			break
		}

		for _, event := range plugin.Events {
			subscriptions[event] = append(subscriptions[event], plugin)
		}
	}

	Session.AddHandler(messageCreate)
	Session.AddHandler(presenceUpdate)

	// We only care about receiving message events.
	Session.Identify.Intents = discordgo.IntentsAll

	// Open a websocket connection to Discord and begin listening.
	err = Session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Close down the Discord session.
	Session.Close()
}

func presenceUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	fmt.Printf("%s(%s)(bot? %t) is now %s\n", m.User.ID, m.User.Username, m.User.Bot, m.Status)
}

func sendChannelMessage(channel string, content string) int32 {
	_, err := Session.ChannelMessageSend(channel, content)
	if err != nil {
		return 1
	}

	return 0
}

func deleteChannelMessage(channel string, id string) int32 {
	err := Session.ChannelMessageDelete(channel, id)
	if err != nil {
		return 1
	}

	return 0
}

// This function will be called every time a new message is
// created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	for _, plugin := range subscriptions[discordgo.IntentsGuildMessages] {
		instance, err := pipeline.GetInstance(plugin.Name, "")

		if err != nil {
			fmt.Println("Error getting plugin instance", "error", err)
		}
		defer instance.Remove()

		err = instance.CallFunction(MessageCreateFunc, nil, m.ChannelID, m.Author.Username, m.ID, m.Content)
		if err != nil {
			fmt.Println("Error calling plugin function", "error", err)
		}
	}
}
