// Copyright 2023 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd || solaris

package vfs

import (
	"io/fs"
	"os"
	"syscall"

	"github.com/cockroachdb/errors"
	"golang.org/x/sys/unix"
)

func wrapOSFileImpl(osFile *os.File) File {
	return &unixFile{File: osFile, fd: osFile.Fd()}
}

func (defaultFS) OpenDir(name string) (File, error) {
	f, err := os.OpenFile(name, syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &unixFile{f, InvalidFd}, nil
}

// Assert that unixFile implements vfs.File.
var _ File = (*unixFile)(nil)

type unixFile struct {
	*os.File
	fd uintptr
}

func (f *unixFile) Stat() (FileInfo, error)                 { return maybeWrapFileInfo(f.File.Stat()) }
func (*unixFile) Prefetch(offset int64, length int64) error { return nil }
func (*unixFile) Preallocate(offset, length int64) error    { return nil }

func (f *unixFile) SyncData() error {
	return f.Sync()
}

func (f *unixFile) SyncTo(int64) (fullSync bool, err error) {
	if err = f.Sync(); err != nil {
		return false, err
	}
	return true, nil
}

func deviceIDFromFileInfo(finfo fs.FileInfo) DeviceID {
	statInfo := finfo.Sys().(*syscall.Stat_t)
	id := DeviceID{
		major: unix.Major(uint64(statInfo.Dev)),
		minor: unix.Minor(uint64(statInfo.Dev)),
	}
	return id
}
