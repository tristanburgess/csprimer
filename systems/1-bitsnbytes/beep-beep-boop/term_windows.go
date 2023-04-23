package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/windows"
)

type termState struct {
	mode uint32
}

func termMakeCBreak(fd int) (*termState, error) {
	var prevState uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &prevState); err != nil {
		return nil, err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		termRestore(&termState{prevState}, STDIN)
		os.Exit(0)
	}()

	newState := prevState &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT)
	// Enable signal handling and ASCII control codes to continue to function
	// https://learn.microsoft.com/en-us/windows/console/high-level-console-modes
	newState |= windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_PROCESSED_OUTPUT
	if err := windows.SetConsoleMode(windows.Handle(fd), newState); err != nil {
		return nil, err
	}

	return &termState{prevState}, nil
}

func termRestore(t *termState, fd int) error {
	return windows.SetConsoleMode(windows.Handle(fd), t.mode)
}
