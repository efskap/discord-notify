package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gen2brain/beeep"
	"github.com/kyoh86/xdg"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var userGuildSettings = make(map[string]discordgo.UserGuildSettings)

var appName = "discord-notify"

func main() {
	var token string
	var soundPath string
	var err error


	flag.StringVar(&token, "t", "", "Discord token")
	flag.StringVar(&soundPath, "sound", "none", `MP3 to play on notifications`)
	flag.Parse()

	if token == "" {
		token, err = readTokenFromFile()
		if err != nil {
			log.Println("error reading token:", err)
			log.Println("put a file called `discord.token` in one of:")
			for _, cfgDir := range xdg.AllConfigDirs() {
				log.Println("  ", cfgDir)
			}
			log.Println("or pass it via the -t flag")
			os.Exit(1)
		}
	}

	if soundPath == "" || strings.ToLower(soundPath) != "none" {
		err = setSound(soundPath)
		if err != nil {
			log.Println("there will be no sound")
			log.Println(err)
			log.Println("you can explicitly disable the notification sound via `-sound none`")
		}
	}

	dg, err := discordgo.New(token)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
	}

	dg.AddHandler(onMessageCreate)
	dg.AddHandler(onReady)
	dg.AddHandler(onGuildSettingsUpdate)

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening Discord session: ", err)
	}

	fmt.Println(appName + " is now running.  Press CTRL-C to exit.")

	// wait until we get a signal to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println("Closing Discord session")
	_ = dg.Close()
}

func shouldShowNotification(s *discordgo.Session, m *discordgo.Message) bool {

	mentionsMe, mentionsRole, mentionsEveryone := mentions(s, m)
	ugs := userGuildSettings[m.GuildID]

	notificationSetting := notifyOption(ugs.MessageNotifications)
	muted := ugs.Muted

	// apply overrides
	for _, override := range ugs.ChannelOverrides {
		if override.ChannelID == m.ChannelID {
			notificationSetting = notifyOption(override.MessageNotifications)
			muted = override.Muted
			break
		}
	}

	// print debug stuff
	if strings.HasPrefix(m.Content, "!testnotify") {
		fmt.Println("----")
		fmt.Println("muted:\t", muted)
		fmt.Println("notificationSetting:\t", notificationSetting)
		fmt.Println()
		fmt.Println("mentionsEveryone:\t", mentionsEveryone)
		fmt.Println("supressEveryone:\t", ugs.SupressEveryone)
		fmt.Println()
		fmt.Println("mentionsRole:\t", mentionsRole)
		fmt.Println("supressRoles:\t", ugs.SupressRoles)
		fmt.Println()
		fmt.Println("mentionsMe:\t", mentionsMe)
		fmt.Println("----")
		return true
	}

	// no need to send the notification when ripcord is focused
	if ripFocused, err := isRipcordFocused(); err == nil && ripFocused {
		return false
	}

	// ignore if we sent the message
	if m.Author.ID == s.State.User.ID {
		return false
	}

	if mentionsEveryone && !ugs.SupressEveryone {
		return true
	}
	if mentionsRole && !ugs.SupressRoles {
		return true
	}
	if mentionsMe {
		return true
	}
	if muted {
		return false
	}
	if notificationSetting == notifyAllMessages {
		return true
	}

	return false
}
func mentions(s *discordgo.Session, m *discordgo.Message) (me, role, everyone bool) {
	myUser := s.State.User
	for _, mentionedUser := range m.Mentions {
		if myUser == mentionedUser {
			me = true
			break
		}
	}

	if guildMember, err := s.GuildMember(m.GuildID, myUser.ID); err != nil {
		// message didn't come from a guild
	} else {
		for _, mentionedRole := range m.MentionRoles {
			for _, guildRole := range guildMember.Roles {
				if mentionedRole == guildRole {
					role = true
					break
				}
			}
		}
	}
	everyone = m.MentionEveryone
	return
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !shouldShowNotification(s, m.Message) {
		return
	}
	playSound()

	authorName := m.Author.String()
	if guildMember, err := s.GuildMember(m.GuildID, m.Author.ID); err == nil {
		authorName = guildMember.Nick
	}
	locationText := "you"
	if channel, err := s.Channel(m.ChannelID); err == nil {
		if channel.GuildID != "" {
			locationText = "#" + channel.Name
		} else if channel.Name != "" {
			locationText = channel.Name
		}
	}
	title := fmt.Sprintf("%s ➔ %s", authorName, locationText)

	if err := beeep.Notify(title, m.Content, getAvatarFor(m.Author)); err != nil {
		fmt.Println("error posting notification:", err)
	}

}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	for _, u := range r.UserGuildSettings {
		userGuildSettings[u.GuildID] = *u
	}
}

func onGuildSettingsUpdate(s *discordgo.Session, u *discordgo.UserGuildSettingsUpdate) {
	userGuildSettings[u.GuildID] = *u.UserGuildSettings
	//fmt.Println("notification settings update")
	//fmt.Printf("%+v\n", *u.UserGuildSettings)
}

type notifyOption int

const (
	notifyAllMessages  notifyOption = 0
	notifyOnlyMentions              = 1
	notifyNever                     = 2
)
