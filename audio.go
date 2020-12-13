package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var buffer *beep.Buffer
var soundName string

func setSoundFromDisk(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error loading sound: %w", err)
	}
	defer f.Close()
	soundName = path
	return setSound(f)
}
func setSoundNone() {
	buffer = nil
}
func setSoundBuiltin(name string) error {
	if name == "" || name == "none" {
		setSoundNone()
		return nil
	}
	f, err := BinFS.Open(filepath.Join("assets", "sounds", name+".mp3"))
	if err != nil {
		return err
	}
	soundName = name
	return setSound(f)
}

func builtInSounds() (results []string) {
	files, err := BinFS.ReadDir("assets/sounds")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp3") {
			results = append(results, strings.TrimSuffix(file.Name(), ".mp3"))
		}
	}
	return results
}

func setSound(data io.ReadCloser) error {
	streamer, format, err := mp3.Decode(data)
	if err != nil {
		return fmt.Errorf("error decoding sound: %w", err)
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}

	buffer = beep.NewBuffer(format)
	buffer.Append(streamer)
	return streamer.Close()
}

func playSound() {
	if buffer != nil {
		sound := buffer.Streamer(0, buffer.Len())
		speaker.Play(sound)
	}
}
