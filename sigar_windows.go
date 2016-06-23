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
	self.List = make([]Cpu, 0, capacity)
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

// Get a list of local filesystems
// Does not apply to SMB volumes
func (self *FileSystemList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	self.List = make([]FileSystem, capacity)

	iter, err := NewWindowsVolumeIterator()
	if err != nil {
		return err
	}

	for iter.Next() {
		volume := iter.Volume()
		self.List = append(self.List, volume)
	}
	iter.Close()

	return iter.Error()
}

func (self *DiskList) Get() error {
	/* Even though these queries are % disk time and ops / sec,
	   we read the raw PDH counter values, not the "cooked" ones.
	   This gives us the underlying number of ticks that would go into
	   computing the cooked metric. */
	diskQueries := []string{
		`\physicaldisk(*)\disk reads/sec`,
		`\physicaldisk(*)\disk read bytes/sec`,
		`\physicaldisk(*)\% disk read time`,
		`\physicaldisk(*)\disk writes/sec`,
		`\physicaldisk(*)\disk write bytes/sec`,
		`\physicaldisk(*)\% disk write time`,
	}

	// Run a PDH query for metrics across all physical disks
	queryResults, err := runRawPdhArrayQueries(diskQueries)
	if err != nil {
		return err
	}

	self.List = make(map[string]DiskIo)
	for disk, counters := range queryResults {
		if disk == "_Total" {
			continue
		}

		self.List[disk] = DiskIo{
			ReadOps:   uint64(counters[0]),
			ReadBytes: uint64(counters[1]),
			// The raw counter for `% disk read time` is measured
			// in 100ns ticks, divide by 10000 to get milliseconds
			ReadTimeMs: uint64(counters[2] / 10000),
			WriteOps:   uint64(counters[3]),
			WriteBytes: uint64(counters[4]),
			// The raw counter for `% disk write time` is measured
			// in 100ns ticks, divide by 10000 to get milliseconds
			WriteTimeMs: uint64(counters[5] / 10000),
			IoTimeMs:    uint64((counters[5] + counters[2]) / 10000),
		}
	}
	return nil
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
	self.Avail = *(*uint64)(unsafe.Pointer(&availableBytes))
	self.Used = self.Total - self.Avail
	return nil
}

func (self *NetIfaceList) Get() error {
	netQueries := []string{
		`\Network Interface(*)\Bytes Sent/sec`,
		`\Network Interface(*)\Bytes Received/sec`,
		`\Network Interface(*)\Packets Sent/sec`,
		`\Network Interface(*)\Packets Received/sec`,
		`\Network Interface(*)\Packets outbound errors`,
		`\Network Interface(*)\Packets received errors`,
		`\Network Interface(*)\Packets outbound discarded`,
		`\Network Interface(*)\Packets received discarded`,
		`\Network Interface(*)\Packets received non-unicast/sec`,
	}
	queryResults, err := runRawPdhArrayQueries(netQueries)
	if err != nil {
		return err
	}
	self.List = make([]NetIface, 0)
	for iface, res := range queryResults {
		ifaceStruct := NetIface {
			Name: iface,
			SendBytes: res[0],
			RecvBytes: res[1],
			SendPackets: res[2],
			RecvPackets: res[3],
			SendErrors: res[4],
			RecvErrors: res[5],
			SendDropped: res[6],
			RecvDropped: res[7],
			RecvMulticast: res[8],
		}
		self.List = append(self.List, ifaceStruct)
	}
	return nil
}

func (self *SystemInfo) Get() error {
	self.Sysname = "Windows"
	return nil
}

func (self *SystemDistribution) Get() error {
	self.Description = "Windows"
	return nil
}

func notImplemented() error {
	panic("Not Implemented")
	return nil
}
