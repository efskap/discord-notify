package main

//go:generate binclude

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/kyoh86/xdg"
	"github.com/lu4p/binclude"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var userGuildSettings = make(map[string]discordgo.UserGuildSettings)

var appName = "discord-notify"
var assetPath = binclude.Include("./assets") // include ./assets with all files and subdirectories
var updateState func(string) = func(s string) { fmt.Println(s) }
var defaultSound = builtInSounds()[0]

func main() {
	var token string
	var useSystray bool
	var soundPath string
	var err error

	flag.StringVar(&token, "t", "", "Discord token")
	flag.BoolVar(&useSystray, "systray", true, "Show an icon in the system tray")
	flag.StringVar(&soundPath, "sound", defaultSound, `MP3 to play on notifications`)
	listSounds := flag.Bool("list-sounds", false, "List available built-in sounds.")
	flag.Parse()
	if listSounds != nil && *listSounds {
		for _, name := range builtInSounds() {
			fmt.Println(name)
		}
		os.Exit(0)
	}

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
		if !strings.ContainsAny(soundPath, "./") && setSoundBuiltin(soundPath) == nil {
			fmt.Println("using built-in sound:", soundPath)
		} else {
			err = setSoundFromDisk(soundPath)
		}
		if err != nil {
			log.Println(err)
			log.Println("notifications will be silent!")
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

	if useSystray {
		systray.AddMenuItem(appName, "").Disable()
		statusMenuItem := systray.AddMenuItem("Logging in...", "")
		statusMenuItem.Disable()
		updateState = func(s string) {
			statusMenuItem.SetTitle(s)
		}
		systray.AddSeparator()
	}
	for {
		err = dg.Open()
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				log.Println("Error connecting:", netErr, "(retrying in a bit)")
				time.Sleep(5 * time.Second)
				continue
			} else {
				log.Fatal("Error opening Discord session: ", err)
			}
		} else {
			break
		}
	}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	if useSystray {
		fmt.Println(appName + " is now running. Press CTRL-C or use the system tray menu to exit.")
		go func() {
			<-ctrlC
			systray.Quit()
		}()
		systray.Run(onSystrayReady, onSystrayExit)
	} else {
		fmt.Println(appName + " is now running. Press CTRL-C to exit.")
		<-ctrlC
	}
	close(ctrlC)
	fmt.Println("Closing Discord session")
	if err = dg.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "Error closing session:", err)
	}
}

func onSystrayReady() {
	systray.SetTitle(appName)
	if err := trySetIcon(); err != nil {
		log.Println("unable to set systray icon:", err)
	}
	type entry struct {
		menu      *systray.MenuItem
		soundName string
	}
	var soundMenus []entry
	mSound := systray.AddMenuItem("Sound", "Temporarily change the sound")
	noneSound := mSound.AddSubMenuItemCheckbox("<none>", "No notification sound", buffer == nil)
	soundMenus = append(soundMenus, entry{
		menu:      noneSound,
		soundName: "",
	})

	for _, name := range builtInSounds() {
		checked := name == soundName
		item := mSound.AddSubMenuItemCheckbox(name, "Built-in sound", checked)
		soundMenus = append(soundMenus, entry{
			menu:      item,
			soundName: name,
		})
	}

	for _, m_ := range soundMenus {
		currentMenu := m_
		go func() {
			for {
				<-currentMenu.menu.ClickedCh
				for _, m := range soundMenus {
					if currentMenu.menu != m.menu {
						m.menu.Uncheck()
					}
				}
				currentMenu.menu.Check()
				setSoundBuiltin(currentMenu.soundName)
			}
		}()
	}

	mQuit := systray.AddMenuItem("Quit", "Quit the app.")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}
func trySetIcon() error {
	f, err := BinFS.ReadFile("assets/icon.ico")
	if err != nil {
		return err
	}
	systray.SetIcon(f)
	return nil
}

func onSystrayExit() {}

func isMe(s *discordgo.Session, u *discordgo.User) bool {
	return u.String() == s.State.User.String()
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
	if strings.HasPrefix(m.Content, "!test") && isMe(s, m.Author) {
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

	for _, mentionedUser := range m.Mentions {
		if isMe(s, mentionedUser) {
			me = true
			break
		}
	}

	if guildMember, err := s.GuildMember(m.GuildID, s.State.User.ID); err != nil {
		// message didn't come from a guild
		role = false
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

	title, body := formatNotification(s, m.Message)

	if err := beeep.Notify(title, body, getAvatarFor(m.Author)); err != nil {
		fmt.Println("error posting notification:", err)
	}

}

func formatNotification(s *discordgo.Session, m *discordgo.Message) (title, body string) {
	authorName := m.Author.String()
	if guildMember, err := s.GuildMember(m.GuildID, m.Author.ID); err == nil && guildMember.Nick != "" {
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
	title = fmt.Sprintf("%s | %s", authorName, locationText)

	var err error
	body, err = m.ContentWithMoreMentionsReplaced(s)
	if err != nil {
		body = m.ContentWithMentionsReplaced()
	}

	// iterate in reverse order since we're adding a prefix each iteration
	for i := len(m.Attachments) - 1; i >= 0; i-- {
		body = fmt.Sprintf("[%s]\n%s", m.Attachments[i].Filename, body)
	}
	return
}

func onReady(_ *discordgo.Session, r *discordgo.Ready) {
	updateState("Logged in as " + r.User.String())
	for _, u := range r.UserGuildSettings {
		userGuildSettings[u.GuildID] = *u
	}
}

func onGuildSettingsUpdate(_ *discordgo.Session, u *discordgo.UserGuildSettingsUpdate) {
	userGuildSettings[u.GuildID] = *u.UserGuildSettings
}

type notifyOption int

const (
	notifyAllMessages  notifyOption = 0
	notifyOnlyMentions              = 1
	notifyNever                     = 2
)
