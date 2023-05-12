//go:build aix || linux || solaris || zos
// +build aix linux solaris zos

package main

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TCGETS
const ioctlWriteTermios = unix.TCSETS
