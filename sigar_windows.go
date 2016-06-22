// Copyright (c) 2012 VMware, Inc.

package sigar

// #cgo CFLAGS: -DPSAPI_VERSION=1
// #cgo LDFLAGS: -lpsapi
// #include <stdlib.h>
// #include <windows.h>
// #include <psapi.h>
import "C"

import (
	"fmt"
	"strconv"
	"syscall"
	"unsafe"
)


const (
	INIT_PID_LIST = 512
	MAX_PID_LIST = 65536

	INIT_MODULE_NAME_SIZE = 128
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

func getPidList() ([]C.DWORD, error) {
	// EnumProcesses doesn't tell you if you've read all the processes,
	// you need to keep making the array bigger until it isn't full
	var sizeRead C.DWORD
	for len := INIT_PID_LIST; len < MAX_PID_LIST; len *= 2 {
		procList := make([]C.DWORD, len)
		listSize := C.DWORD(unsafe.Sizeof(procList[0])*uintptr(len))
		success := C.EnumProcesses(&procList[0], listSize, &sizeRead)
		if success == C.FALSE {
			return nil, fmt.Errorf("Failed to enumerate list of processes")
		}
		// If we've read less than a full array, slice the list down to just the relevant entries
		if sizeRead < listSize {
			return procList[:uintptr(sizeRead)/(unsafe.Sizeof(procList[0]))], nil
		}
	}
	return nil, fmt.Errorf("Expanded process list up to %v, array was full", MAX_PID_LIST)
}	

func (self *ProcList) Get() error {
	procList, err := getPidList()
	if err != nil {
		return err
	}
	self.List = make([]int, len(procList))
	for i := range procList {
		self.List[i] = int(procList[i])
	}
	return nil
}

func openProcess(pid int) C.HANDLE {
	return C.OpenProcess(C.PROCESS_QUERY_INFORMATION | C.PROCESS_VM_READ, C.FALSE, C.DWORD(pid))
}

func (self *ProcState) Get(pid int) error {
	procH := openProcess(pid)
	if uintptr(unsafe.Pointer(procH)) == 0 {
		return fmt.Errorf("Unable to open process %v: error code %v", pid, C.GetLastError())
	}
	defer C.CloseHandle(procH)

	// Get the name of only the first module
	var firstModule C.HMODULE
	var size C.DWORD
	if C.EnumProcessModules(procH, &firstModule, C.DWORD(unsafe.Sizeof(firstModule)), &size) != C.TRUE {
		return fmt.Errorf("Unable to get name of process image for pid %v: error code %v", pid, C.GetLastError())
	}

	imageName := make([]uint8, INIT_MODULE_NAME_SIZE)
	imageNameLen := C.GetModuleBaseName(procH, firstModule, (*C.CHAR)(unsafe.Pointer(&imageName[0])), INIT_MODULE_NAME_SIZE)
	if imageNameLen == 0 {
		return fmt.Errorf("Unable to get name of first module name for pid %v: error code %v", pid, C.GetLastError())
	}
	self.Name = C.GoString((*C.char)(unsafe.Pointer(&imageName[0])))

	// Get process priority
	self.Priority = int(C.GetPriorityClass(procH))
	if self.Priority == 0 {
		return fmt.Errorf("Unable to get process priority for pid %v: error code %v", pid, C.GetLastError())
	}

	return nil
}

func (self *ProcMem) Get(pid int) error {
	procH := openProcess(pid)
	if uintptr(unsafe.Pointer(procH)) == 0 {
		return fmt.Errorf("Unable to open process %v: error code %v", pid, C.GetLastError())
	}
	defer C.CloseHandle(procH)

	var counters C.PROCESS_MEMORY_COUNTERS
	if C.GetProcessMemoryInfo(procH, &counters, C.DWORD(unsafe.Sizeof(counters))) != C.TRUE {
		return fmt.Errorf("Unable to get memory info for process %v: error code %v", pid, C.GetLastError())
	}
	self.Resident = uint64(counters.PagefileUsage)
	self.PageFaults = uint64(counters.PageFaultCount)
	self.Size = uint64(counters.WorkingSetSize)
	return nil
}

func (self *ProcTime) Get(pid int) error {
	procH := openProcess(pid)
	if uintptr(unsafe.Pointer(procH)) == 0 {
		return fmt.Errorf("Unable to open process %v: error code %v", pid, C.GetLastError())
	}
	defer C.CloseHandle(procH)

	var creationTime syscall.Filetime
	var exitTime syscall.Filetime
	var kernelTime syscall.Filetime
	var userTime syscall.Filetime
	if C.GetProcessTimes(procH, (*C.FILETIME)(unsafe.Pointer(&creationTime)), (*C.FILETIME)(unsafe.Pointer(&exitTime)), (*C.FILETIME)(unsafe.Pointer(&kernelTime)), (*C.FILETIME)(unsafe.Pointer(&userTime))) != C.TRUE {
		return fmt.Errorf("Unable to get process time for pid %v: error code %v", pid, C.GetLastError())
	}
	// Convert the FILETIME to nanos, then divide to get millis
	self.StartTime = uint64(creationTime.Nanoseconds()) / 1000000

	// Convert the 100-nanosecond ticks to millis
	self.User = (uint64(userTime.HighDateTime) << 32 + uint64(userTime.LowDateTime)) / 10000
	self.Sys = (uint64(kernelTime.HighDateTime) << 32 + uint64(kernelTime.LowDateTime)) / 10000
	self.Total = self.User + self.Sys
	return nil
}

func (self *ProcArgs) Get(pid int) error {
	return notImplemented()
}

func (self *ProcExe) Get(pid int) error {
	procH := openProcess(pid)
	if uintptr(unsafe.Pointer(procH)) == 0 {
		return fmt.Errorf("Unable to open process %v: error code %v", pid, C.GetLastError())
	}
	defer C.CloseHandle(procH)

	// Get the exe for the process image
	imageName := make([]uint8, INIT_MODULE_NAME_SIZE)
	imageNameLen := C.GetProcessImageFileName(procH, (*C.CHAR)(unsafe.Pointer(&imageName[0])), INIT_MODULE_NAME_SIZE)
	if imageNameLen == 0 {
		return fmt.Errorf("Unable to get name of process image for pid %v: error code %v", pid, C.GetLastError())
	}
	self.Name = C.GoString((*C.char)(unsafe.Pointer(&imageName[0])))
	return nil
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
