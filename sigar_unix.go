// Copyright (c) 2012 VMware, Inc.

// +build darwin freebsd linux netbsd openbsd

package sigar

import "syscall"

func (self *FileSystemUsage) Get(path string) error {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return err
	}

	bsize := stat.Bsize / 512

	self.Total = (uint64(stat.Blocks) * uint64(bsize)) >> 1
	self.Free = (uint64(stat.Bfree) * uint64(bsize)) >> 1
	self.Avail = (uint64(stat.Bavail) * uint64(bsize)) >> 1
	self.Used = self.Total - self.Free
	self.Files = stat.Files
	self.FreeFiles = stat.Ffree

	return nil
}

func (self *SystemInfo) Get() error {
	var uname syscall.Utsname
	err := syscall.Uname(&uname)
	if err != nil {
		return err
	}
	self.Sysname = bytePtrToString(&uname.Sysname[0])
	self.Nodename = bytePtrToString(&uname.Nodename[0])
	self.Release = bytePtrToString(&uname.Release[0])
	self.Version = bytePtrToString(&uname.Version[0])
	self.Machine = bytePtrToString(&uname.Machine[0])
	self.Domainname = bytePtrToString(&uname.Domainname[0])

	return nil
}
