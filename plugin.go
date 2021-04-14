package main

import "github.com/bwmarrin/discordgo"

type Plugin struct {
	Name    string
	Version string
	Path    string
	Events  []discordgo.Intent
}

func (p *Plugin) String() string {
	return p.Name + p.Version
}
