//+build !windows

package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"strings"
)

func isRipcordFocused() (bool, error) {
	if wmClasses, err := activeWindowClasses(); err != nil {
		return false, err
	} else {
		for _, wmClass := range wmClasses {
			if wmClass == "Ripcord" {
				return true, nil
			}
		}
	}
	return false, nil
}

// reuse connection
var xConn *xgb.Conn

// adapted from https://github.com/BurntSushi/xgb/blob/master/examples/get-active-window/main.go
func activeWindowClasses() ([]string, error) {
	if xConn == nil {
		var err error
		xConn, err = xgb.NewConn()
		if err != nil {
			return nil, fmt.Errorf("couldn't connect to X: %w", err)
		}
	}

	// Get the window id of the root window.
	setup := xproto.Setup(xConn)
	root := setup.DefaultScreen(xConn).Root

	// Get the atom id (i.e., intern an atom) of "_NET_ACTIVE_WINDOW".
	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(xConn, true, uint16(len(aname)),
		aname).Reply()
	if err != nil {
		return nil, fmt.Errorf("couldn't get atom of %s: %w", aname, err)
	}

	// Get the atom id (i.e., intern an atom) of "_NET_WM_NAME".
	aname = "WM_CLASS"
	classAtom, err := xproto.InternAtom(xConn, true, uint16(len(aname)),
		aname).Reply()
	if err != nil {
		return nil, fmt.Errorf("couldn't get atom of %s: %w", aname, err)
	}

	// Get the actual value of _NET_ACTIVE_WINDOW.
	// Note that 'reply.Value' is just a slice of bytes, so we use an
	// XGB helper function, 'Get32', to pull an unsigned 32-bit integer out
	// of the byte slice. We then convert it to an X resource id so it can
	// be used to get the name of the window in the next GetProperty request.
	reply, err := xproto.GetProperty(xConn, false, root, activeAtom.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, fmt.Errorf("couldn't get active window: %w", err)
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))

	// Now get the value of _NET_WM_NAME for the active window.
	// Note that this time, we simply convert the resulting byte slice,
	// reply.Value, to a string.
	reply, err = xproto.GetProperty(xConn, false, windowId, classAtom.Atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, fmt.Errorf("couldn't get active window class: %w", err)
	}

	// extract null terminated strings from byte buffer
	classes := strings.Split(string(reply.Value), "\000")
	// last one is empty because the buffer ends with null
	classes = classes[:len(classes)-1]

	return classes, nil
}
