package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)

var buffer *beep.Buffer

func setSound(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error loading sound: %w", err)
	}
	streamer, format, err := mp3.Decode(f)
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