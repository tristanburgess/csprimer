// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"golang.org/x/sys/windows"
)

type termState struct {
	mode uint32
}

func termMakeCBreak(fd int) (*termState, error) {
	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return nil, err
	}
	newMode := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT)
	// Enable signal handling and ASCII control codes to continue to function
	// https://learn.microsoft.com/en-us/windows/console/high-level-console-modes
	newMode |= windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_PROCESSED_OUTPUT
	if err := windows.SetConsoleMode(windows.Handle(fd), newMode); err != nil {
		return nil, err
	}
	return &termState{st}, nil
}

func restore(fd int, s *termState) error {
	return windows.SetConsoleMode(windows.Handle(fd), s.mode)
}
