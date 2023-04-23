//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package main

import (
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

	oldState := termState{termios: *termios}

	// https://linux.die.net/man/3/cbreak
	termios.Iflag &^= unix.ECHO | unix.ICANON
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return nil, err
	}

	return &oldState, nil
}

func restore(fd int, s *termState) error {
	return unix.IoctlSetTermios(fd, ioctlWriteTermios, &s.termios)
}
