// +build !gen

//go:generate go run -tags=gen .
package main

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

const appName = "discord-notify"
const appHomepage = "https://github.com/efskap/discord-notify"

var assetPath = binclude.Include("assets") // include ./assets with all files and subdirectories
var updateState = func(s ...string) { fmt.Println(strings.Join(s, " ")) }

var exitSignal = make(chan bool, 1)

func main() {
	var token string
	var useSystray bool
	var soundPath string
	var err error

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	flag.StringVar(&token, "t", "", "Discord token")
	flag.BoolVar(&useSystray, "systray", true, "Show an icon in the system tray")
	flag.StringVar(&soundPath, "sound", builtInSounds()[0], `MP3 to play on notifications`)
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

		if !strings.ContainsAny(soundPath, "./") {
			err = setSoundBuiltin(soundPath)

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
	dg.AddHandler(onDisconnect)

	defer func() {
		fmt.Println("Closing Discord session")
		if err = dg.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing session:", err)
		}
	}()

	if useSystray {

		statusMenuItem := systray.AddMenuItem("Logging in...", "")
		statusMenuItem.Disable()
		systray.AddSeparator()
		oldUpdateState := updateState
		updateState = func(s ...string) {
			oldUpdateState(s...)
			statusMenuItem.SetTitle(strings.Join(s, " "))
		}

		defer systray.Quit()
		go systray.Run(onSystrayReady, onSystrayExit)
	}
	for {
		err = dg.Open()
		if err != nil {
			setStatusIcon(true)
			var netErr net.Error
			if errors.As(err, &netErr) {
				updateState(fmt.Sprintln("Error connecting:", netErr, "(retrying in a bit)"))
				time.Sleep(5 * time.Second)
				continue
			} else {
				updateState(fmt.Errorf("error opening Discord session: %w", err).Error())
				if useSystray {
					break
				} else {
					return
				}
			}
		} else {
			break
		}
	}

	if useSystray {
		fmt.Println(appName + " is now running. Press CTRL-C or use the system tray menu to exit.")
	} else {
		fmt.Println(appName + " is now running. Press CTRL-C to exit.")
	}
	select {
	case <-exitSignal:
		break
	case <-ctrlC:
		break
	}
}
func setStatusIcon(isError bool) {
	iconPath := "assets/icon.ico"
	if isError {
		iconPath = "assets/icon_err.ico"
	}
	if err := trySetIcon(iconPath); err != nil {
		log.Println("unable to set systray icon:", err)
	}

}
func onSystrayReady() {
	systray.SetTitle(appName)
	setStatusIcon(false)
	type entry struct {
		menu      *systray.MenuItem
		soundName string
	}
	var soundMenus []entry
	mSound := systray.AddMenuItem("Sound", "Temporarily change the sound")
	noneSound := mSound.AddSubMenuItemCheckbox("<none>", "No notification sound", currentSound.buffer == nil)
	soundMenus = append(soundMenus, entry{
		menu:      noneSound,
		soundName: "",
	})

	for _, name := range builtInSounds() {
		checked := name == currentSound.name
		item := mSound.AddSubMenuItemCheckbox(name, "Built-in sound", checked)
		soundMenus = append(soundMenus, entry{
			menu:      item,
			soundName: name,
		})
	}

	for _, m_ := range soundMenus {
		currentMenu := m_
		go func() {
			for range currentMenu.menu.ClickedCh {
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

	m := systray.AddMenuItem("Open GitHub", "Visit repo")
	go func() {
		for range m.ClickedCh {
			if err := openBrowser(appHomepage); err != nil {
				fmt.Fprintln(os.Stderr, "error opening", appHomepage, ":", err)
			}
		}
	}()
	mQuit := systray.AddMenuItem("Quit", "Quit the app.")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

var curIcon = ""

func trySetIcon(path string) error {
	if path == curIcon {
		return nil
	}
	curIcon = path
	f, err := BinFS.ReadFile(path)
	if err != nil {
		return err
	}
	systray.SetIcon(f)
	return nil
}

func onSystrayExit() {
	exitSignal <- true
}

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
	setStatusIcon(false)
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
	if guildMember, err := s.GuildMember(m.GuildID, m.Author.ID); err == nil {
		if guildMember.Nick != "" {
			authorName = guildMember.Nick
		} else {
			authorName = guildMember.User.Username
		}
	}
	locationText := "you"
	if channel, err := s.Channel(m.ChannelID); err == nil {
		if channel.GuildID != "" {
			locationText = "#" + channel.Name
		} else if channel.Name != "" {
			locationText = channel.Name
		}
	}
	title = fmt.Sprintf("%s (%s)", authorName, locationText)

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
	setStatusIcon(false)
	updateState("Logged in as " + r.User.String())
	for _, u := range r.UserGuildSettings {
		userGuildSettings[u.GuildID] = *u
	}
}

func onGuildSettingsUpdate(_ *discordgo.Session, u *discordgo.UserGuildSettingsUpdate) {
	userGuildSettings[u.GuildID] = *u.UserGuildSettings
}

func onDisconnect(_ *discordgo.Session, d *discordgo.Disconnect) {
	updateState("Disconnected!")
	setStatusIcon(true)
}

type notifyOption int

const (
	notifyAllMessages  notifyOption = 0
	notifyOnlyMentions              = 1
	notifyNever                     = 2
)
