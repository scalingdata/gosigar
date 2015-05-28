// Copyright (c) 2012 VMware, Inc.

package sigar

// #include <stdlib.h>
// #include <windows.h>
import "C"

import (
	"fmt"
	"strconv"
	"unsafe"
)

func init() {
}

func (self *LoadAverage) Get() error {
	return nil
}

func (self *Uptime) Get() error {
	return nil
}

func (self *Mem) Get() error {
	var statex C.MEMORYSTATUSEX
	statex.dwLength = C.DWORD(unsafe.Sizeof(statex))

	succeeded := C.GlobalMemoryStatusEx(&statex)
	if succeeded == C.FALSE {
		lastError := C.GetLastError()
		return fmt.Errorf("GlobalMemoryStatusEx failed with error: %d", int(lastError))
	}

	self.Total = uint64(statex.ullTotalPhys)
	self.Free = uint64(statex.ullAvailPhys)
	self.Used = self.Total - self.Free
	return nil
}

func (self *Swap) Get() error {
	swapQueries := []string{
		`\memory\committed bytes`,
		`\memory\commit limit`,
	}

	queryResults, err := runRawPdhQueries(swapQueries)
	if err != nil {
		return err
	}
	self.Used = uint64(queryResults[0])
	self.Total = uint64(queryResults[1])
	self.Free = self.Total - self.Used
	return nil
}

func (self *Cpu) Get() error {
	cpuQueries := []string{
		`\processor(_Total)\% idle time`,
		`\processor(_Total)\% user time`,
		`\processor(_Total)\% privileged time`,
		`\processor(_Total)\% interrupt time`,
	}
	queryResults, err := runRawPdhQueries(cpuQueries)
	if err != nil {
		return err
	}

	self.populateFromPdh(queryResults)
	return nil
}

func (self *CpuList) Get() error {
	cpuQueries := []string{
		`\processor(*)\% idle time`,
		`\processor(*)\% user time`,
		`\processor(*)\% privileged time`,
		`\processor(*)\% interrupt time`,
	}
	// Run a PDH query for all CPU metrics
	queryResults, err := runRawPdhArrayQueries(cpuQueries)
	if err != nil {
		return err
	}
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	self.List = make([]Cpu, capacity)
	for cpu, counters := range queryResults {
		index := 0
		if cpu == "_Total" {
			continue
		}

		index, err := strconv.Atoi(cpu)
		if err != nil {
			continue
		}

		// Expand the array to accomodate this CPU id
		for i := len(self.List); i <= index; i++ {
			self.List = append(self.List, Cpu{})
		}

		// Populate the relevant fields
		self.List[index].populateFromPdh(counters)
	}
	return nil
}

func (self *Cpu) populateFromPdh(values []uint64) {
	self.Idle = values[0]
	self.User = values[1]
	self.Sys = values[2]
	self.Irq = values[3]
}

func (self *FileSystemList) Get() error {
	return notImplemented()
}

func (self *DiskList) Get() error {
	return notImplemented()
}

func (self *ProcList) Get() error {
	return notImplemented()
}

func (self *ProcState) Get(pid int) error {
	return notImplemented()
}

func (self *ProcMem) Get(pid int) error {
	return notImplemented()
}

func (self *ProcTime) Get(pid int) error {
	return notImplemented()
}

func (self *ProcArgs) Get(pid int) error {
	return notImplemented()
}

func (self *ProcExe) Get(pid int) error {
	return notImplemented()
}

func (self *FileSystemUsage) Get(path string) error {
	var availableBytes C.ULARGE_INTEGER
	var totalBytes C.ULARGE_INTEGER
	var totalFreeBytes C.ULARGE_INTEGER

	pathChars := C.CString(path)
	defer C.free(unsafe.Pointer(pathChars))

	succeeded := C.GetDiskFreeSpaceEx((*C.CHAR)(pathChars), &availableBytes, &totalBytes, &totalFreeBytes)
	if succeeded == C.FALSE {
		lastError := C.GetLastError()
		return fmt.Errorf("GetDiskFreeSpaceEx failed with error: %d", int(lastError))
	}

	self.Total = *(*uint64)(unsafe.Pointer(&totalBytes))
	return nil
}

func notImplemented() error {
	panic("Not Implemented")
	return nil
}
