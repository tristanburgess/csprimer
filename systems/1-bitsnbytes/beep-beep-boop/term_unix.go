//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

type termState struct {
	termios unix.Termios
}

func termMakeCBreak(fd int) (*termState, error) {
	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return nil, err
	}

	prevState := termState{termios: *termios}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		termRestore(&prevState, STDIN)
		os.Exit(0)
	}()

	// https://linux.die.net/man/3/cbreak
	termios.Lflag &^= unix.ECHO | unix.ICANON
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return nil, err
	}

	return &prevState, nil
}

func termRestore(t *termState, fd int) error {
	return unix.IoctlSetTermios(fd, ioctlWriteTermios, &t.termios)
}
