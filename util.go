package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/kyoh86/xdg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// no error handling because a blank avatar isn't fatal
func getAvatarFor(u *discordgo.User) (avatarPath string) {
	avatarDir := filepath.Join(xdg.CacheHome(), "discord-avatars")
	if err := os.MkdirAll(avatarDir, os.ModePerm); err != nil {
		log.Println("warning: unable to make avatar cache dir: ", err)
		return
	}
	avatarPath = filepath.Join(avatarDir, u.Avatar+".png")
	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		resp, err := http.Get(u.AvatarURL("64"))
		if err != nil {
			log.Println("warning: unable to download avatar: ", err)
			return
		}
		defer resp.Body.Close()
		outFile, err := os.Create(avatarPath)
		if err != nil {
			log.Println("warning: ", err)
			return
		}
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			log.Println("warning: unable to save avatar: ", err)
			return
		}
	}
	return avatarPath
}

func readTokenFromFile() (string, error) {
	configPath, err := xdg.FindConfigFile("discord.token")
	if err != nil {
		return "", err
	}
	tokenBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(tokenBytes))
	return token, nil
}

func openBrowser(url string) (err error) {
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return
}
