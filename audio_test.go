package main

import (
	"testing"
)

func Test_builtInSounds(t *testing.T) {
	if len(builtInSounds()) < 1 {
		t.Fatal("no built in sounds found")
	}

}

func Test_setSoundBuiltin(t *testing.T) {
	for _, builtinSound := range builtInSounds() {
		t.Run(builtinSound, func(t *testing.T) {
			if err := setSoundBuiltin(builtinSound); err != nil {
				t.Errorf("setSoundBuiltin() error = %v", err)
			}
		})
	}
}

func Test_setSoundNone(t *testing.T) {
	Test_builtInSounds(t)
	if err := setSoundBuiltin(builtInSounds()[0]); err != nil {
		t.Fatal(err)
	}
	if buffer == nil {
		t.Fatal("buffer already nil")
	}
	setSoundNone()
	if buffer != nil {
		t.Fatal("buffer not nil after setSoundNull")
	}

}
