// Copyright (c) 2012 VMware, Inc.

package sigar

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var system struct {
	ticks uint64
	btime uint64
}

var Procd string

func init() {
	system.ticks = 100 // C.sysconf(C._SC_CLK_TCK)

	Procd = "/proc"

	// grab system boot time
	readFile(Procd+"/stat", func(line string) bool {
		if strings.HasPrefix(line, "btime") {
			system.btime, _ = strtoull(line[6:])
			return false // stop reading
		}
		return true
	})
}

func (self *LoadAverage) Get() error {
	line, err := ioutil.ReadFile(Procd + "/loadavg")
	if err != nil {
		return nil
	}

	fields := strings.Fields(string(line))

	self.One, _ = strconv.ParseFloat(fields[0], 64)
	self.Five, _ = strconv.ParseFloat(fields[1], 64)
	self.Fifteen, _ = strconv.ParseFloat(fields[2], 64)

	return nil
}

func (self *Uptime) Get() error {
	sysinfo := syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(&sysinfo); err != nil {
		return err
	}

	self.Length = float64(sysinfo.Uptime)

	return nil
}

func (self *Mem) Get() error {
	var buffers, cached uint64
	table := map[string]*uint64{
		"MemTotal": &self.Total,
		"MemFree":  &self.Free,
		"Buffers":  &buffers,
		"Cached":   &cached,
	}

	if err := parseMeminfo(table); err != nil {
		return err
	}

	self.Used = self.Total - self.Free
	kern := buffers + cached
	self.ActualFree = self.Free + kern
	self.ActualUsed = self.Used - kern

	return nil
}

func (self *Swap) Get() error {
	table := map[string]*uint64{
		"SwapTotal": &self.Total,
		"SwapFree":  &self.Free,
	}

	if err := parseMeminfo(table); err != nil {
		return err
	}

	self.Used = self.Total - self.Free
	return nil
}

func (self *Cpu) Get() error {
	return readFile(Procd+"/stat", func(line string) bool {
		if len(line) > 4 && line[0:4] == "cpu " {
			parseCpuStat(self, line)
			return false
		}
		return true

	})
}

func (self *CpuList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 4
	}
	list := make([]Cpu, 0, capacity)

	err := readFile(Procd+"/stat", func(line string) bool {
		if len(line) > 3 && line[0:3] == "cpu" && line[3] != ' ' {
			cpu := Cpu{}
			parseCpuStat(&cpu, line)
			list = append(list, cpu)
		}
		return true
	})

	self.List = list

	return err
}

func (self *FileSystemList) Get() error {
	capacity := len(self.List)
	if capacity == 0 {
		capacity = 10
	}
	fslist := make([]FileSystem, 0, capacity)

	err := readFile("/etc/mtab", func(line string) bool {
		fields := strings.Fields(line)

		fs := FileSystem{}
		fs.DevName = fields[0]
		fs.DirName = fields[1]
		fs.SysTypeName = fields[2]
		fs.Options = fields[3]

		fslist = append(fslist, fs)

		return true
	})

	self.List = fslist

	return err
}

func (self *DiskList) Get() error {
	/* List all the partitions, and check the major/minor device ID
	   to find which are devices vs. partitions (ex. sda v. sda1) */
	devices := make(map[string]bool)
	diskList := make(map[string]DiskIo)
	err := readFile(Procd+"/partitions", func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			return true
		}
		majorDevId, err := strtoull(fields[0])
		if err != nil {
			return true
		}
		minorDevId, err := strtoull(fields[1])
		if err != nil {
			return true
		}
		if isNotPartition(majorDevId, minorDevId) {
			devices[fields[3]] = true
		}
		return true
	})

	/* Get all device stats from /proc/diskstats and filter by
	   devices from /proc/partitions */
	err = readFile(Procd+"/diskstats", func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			return true
		}
		deviceName := fields[2]
		if _, ok := devices[deviceName]; !ok {
			return true
		}
		io := DiskIo {}
		io.ReadOps, _ = strtoull(fields[3])
		readBytes, _ := strtoull(fields[5])
		io.ReadBytes = readBytes * 512
		io.ReadTimeMs, _ = strtoull(fields[6])
		io.WriteOps, _ = strtoull(fields[7])
		writeBytes, _ := strtoull(fields[9])
		io.WriteBytes = writeBytes * 512
		io.WriteTimeMs, _ = strtoull(fields[10])
		io.IoTimeMs, _ = strtoull(fields[12])
		diskList[deviceName] = io
		return true
	})
	self.List = diskList
	return err
}

func (self *ProcList) Get() error {
	dir, err := os.Open(Procd)
	if err != nil {
		return err
	}
	defer dir.Close()

	const readAllDirnames = -1 // see os.File.Readdirnames doc

	names, err := dir.Readdirnames(readAllDirnames)
	if err != nil {
		return err
	}

	capacity := len(names)
	list := make([]int, 0, capacity)

	for _, name := range names {
		if name[0] < '0' || name[0] > '9' {
			continue
		}
		pid, err := strconv.Atoi(name)
		if err == nil {
			list = append(list, pid)
		}
	}

	self.List = list

	return nil
}

func (self *ProcState) Get(pid int) error {
	contents, err := readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	self.Name = fields[1][1 : len(fields[1])-1] // strip ()'s

	self.State = RunState(fields[2][0])

	self.Ppid, _ = strconv.Atoi(fields[3])

	self.Tty, _ = strconv.Atoi(fields[6])

	self.Priority, _ = strconv.Atoi(fields[17])

	self.Nice, _ = strconv.Atoi(fields[18])

	self.Processor, _ = strconv.Atoi(fields[38])

	return nil
}

func (self *ProcMem) Get(pid int) error {
	contents, err := readProcFile(pid, "statm")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	size, _ := strtoull(fields[0])
	self.Size = size << 12

	rss, _ := strtoull(fields[1])
	self.Resident = rss << 12

	share, _ := strtoull(fields[2])
	self.Share = share << 12

	contents, err = readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields = strings.Fields(string(contents))

	self.MinorFaults, _ = strtoull(fields[10])
	self.MajorFaults, _ = strtoull(fields[12])
	self.PageFaults = self.MinorFaults + self.MajorFaults

	return nil
}

func (self *ProcTime) Get(pid int) error {
	contents, err := readProcFile(pid, "stat")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))

	user, _ := strtoull(fields[13])
	sys, _ := strtoull(fields[14])
	// convert to millis
	self.User = user * (1000 / system.ticks)
	self.Sys = sys * (1000 / system.ticks)
	self.Total = self.User + self.Sys

	// convert to millis
	self.StartTime, _ = strtoull(fields[21])
	self.StartTime /= system.ticks
	self.StartTime += system.btime
	self.StartTime *= 1000

	return nil
}

func (self *ProcArgs) Get(pid int) error {
	contents, err := readProcFile(pid, "cmdline")
	if err != nil {
		return err
	}

	bbuf := bytes.NewBuffer(contents)

	var args []string

	for {
		arg, err := bbuf.ReadBytes(0)
		if err == io.EOF {
			break
		}
		args = append(args, string(chop(arg)))
	}

	self.List = args

	return nil
}

func (self *ProcExe) Get(pid int) error {
	fields := map[string]*string{
		"exe":  &self.Name,
		"cwd":  &self.Cwd,
		"root": &self.Root,
	}

	for name, field := range fields {
		val, err := os.Readlink(procFileName(pid, name))

		if err != nil {
			return err
		}

		*field = val
	}

	return nil
}

func parseMeminfo(table map[string]*uint64) error {
	return readFile(Procd+"/meminfo", func(line string) bool {
		fields := strings.Split(line, ":")

		if ptr := table[fields[0]]; ptr != nil {
			num := strings.TrimLeft(fields[1], " ")
			val, err := strtoull(strings.Fields(num)[0])
			if err == nil {
				*ptr = val * 1024
			}
		}

		return true
	})
}

func parseCpuStat(self *Cpu, line string) error {
	fields := strings.Fields(line)

	self.User, _ = strtoull(fields[1])
	self.Nice, _ = strtoull(fields[2])
	self.Sys, _ = strtoull(fields[3])
	self.Idle, _ = strtoull(fields[4])
	self.Wait, _ = strtoull(fields[5])
	self.Irq, _ = strtoull(fields[6])
	self.SoftIrq, _ = strtoull(fields[7])
	self.Stolen, _ = strtoull(fields[8])
	/* Guest was added in 2.6, not available on all kernels */
	if len(fields) > 9 {
	  self.Guest, _ = strtoull(fields[9])
	}

	return nil
}

func readFile(file string, handler func(string) bool) error {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(bytes.NewBuffer(contents))

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if !handler(string(line)) {
			break
		}
	}

	return nil
}

func strtoull(val string) (uint64, error) {
	return strconv.ParseUint(val, 10, 64)
}

func procFileName(pid int, name string) string {
	return Procd + "/" + strconv.Itoa(pid) + "/" + name
}

func readProcFile(pid int, name string) ([]byte, error) {
	path := procFileName(pid, name)
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			if perr.Err == syscall.ENOENT {
				return nil, syscall.ESRCH
			}
		}
	}

	return contents, err
}

/* For SCSI and IDE devices, only display devices and not individual partitions.
   For other major numbers, show all devices regardless of minor (for LVM, for example).
   As described here: http://www.linux-tutorial.info/modules.php?name=MContent&pageid=94 */
func isNotPartition(majorDevId, minorDevId uint64) bool {
	if majorDevId == 3 || majorDevId == 22 ||        // IDE0_MAJOR IDE1_MAJOR
		majorDevId == 33 || majorDevId == 34 ||  // IDE2_MAJOR IDE3_MAJOR
		majorDevId == 56 || majorDevId == 57 ||  // IDE4_MAJOR IDE5_MAJOR
		(majorDevId >= 88 && majorDevId <= 91) { // IDE6_MAJOR to IDE_IDE9_MAJOR
		return (minorDevId & 0x3F) == 0 // IDE uses bottom 10 bits for partitions
	}
	if majorDevId == 8 || // SCSI_DISK0_MAJOR 
		(majorDevId >= 65 && majorDevId <= 71) ||  // SCSI_DISK1_MAJOR to SCSI_DISK7_MAJOR
		(majorDevId >= 128 && majorDevId <= 135) { // SCSI_DISK8_MAJOR to SCSI_DISK15_MAJOR
		return (minorDevId & 0x0F) == 0 // SCSI uses bottom 8 bits for partitions
	}
	return true
}

func (self *SystemDistribution) Get() error {
	b, err := ioutil.ReadFile("/etc/redhat-release")
	if err != nil {
		return err
	}

	self.Description = string(b)
	return nil
}
