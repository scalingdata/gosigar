package sigar

import (
	"time"
)

type Sigar interface {
	CollectCpuStats(collectionInterval time.Duration) (<-chan Cpu, chan<- struct{})
	GetLoadAverage() (LoadAverage, error)
	GetMem() (Mem, error)
	GetSwap() (Swap, error)
	GetFileSystemUsage(string) (FileSystemUsage, error)
	GetSystemInfo() (SystemInfo, error)
	GetSystemDistribution() (SystemDistribution, error)
}

type Cpu struct {
	User    uint64
	Nice    uint64
	Sys     uint64
	Idle    uint64
	Wait    uint64
	Irq     uint64
	SoftIrq uint64
	Stolen  uint64
	Guest   uint64
}

func (cpu *Cpu) Total() uint64 {
	return cpu.User + cpu.Nice + cpu.Sys + cpu.Idle +
		cpu.Wait + cpu.Irq + cpu.SoftIrq + cpu.Stolen + cpu.Guest
}

func (cpu Cpu) Delta(other Cpu) Cpu {
	return Cpu{
		User:    cpu.User - other.User,
		Nice:    cpu.Nice - other.Nice,
		Sys:     cpu.Sys - other.Sys,
		Idle:    cpu.Idle - other.Idle,
		Wait:    cpu.Wait - other.Wait,
		Irq:     cpu.Irq - other.Irq,
		SoftIrq: cpu.SoftIrq - other.SoftIrq,
		Stolen:  cpu.Stolen - other.Stolen,
		Guest:   cpu.Guest - other.Guest,
	}
}

type LoadAverage struct {
	One, Five, Fifteen float64
}

type Uptime struct {
	Length float64
}

type Mem struct {
	Total      uint64
	Used       uint64
	Free       uint64
	ActualFree uint64
	ActualUsed uint64
}

type Swap struct {
	Total uint64
	Used  uint64
	Free  uint64
}

type CpuList struct {
	List []Cpu
}

type FileSystem struct {
	DirName     string
	DevName     string
	TypeName    string
	SysTypeName string
	Options     string
	Flags       uint32
}

type FileSystemList struct {
	List []FileSystem
}

type FileSystemUsage struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Avail     uint64
	Files     uint64
	FreeFiles uint64
}

type NetIface struct {
	Name       string
	MTU        uint64
	Mac        string
	LinkStatus string

	SendBytes      uint64
	RecvBytes      uint64
	SendPackets    uint64
	RecvPackets    uint64
	SendCompressed uint64
	RecvCompressed uint64
	RecvMulticast  uint64

	SendErrors     uint64
	RecvErrors     uint64
	SendDropped    uint64
	RecvDropped    uint64
	SendFifoErrors uint64
	RecvFifoErrors uint64

	RecvFramingErrors uint64
	SendCarrier       uint64
	SendCollisions    uint64
}

type NetIfaceList struct {
	List []NetIface
}

type ProcList struct {
	List []int
}

type RunState byte

const (
	RunStateSleep   = 'S'
	RunStateRun     = 'R'
	RunStateStop    = 'T'
	RunStateZombie  = 'Z'
	RunStateIdle    = 'D'
	RunStateUnknown = '?'
)

type ProcState struct {
	Name      string
	State     RunState
	Ppid      int
	Tty       int
	Priority  int
	Nice      int
	Processor int
}

type ProcIo struct {
	ReadBytes  uint64
	WriteBytes uint64
	ReadOps    uint64
	WriteOps   uint64
}

type ProcMem struct {
	Size        uint64
	Resident    uint64
	Share       uint64
	MinorFaults uint64
	MajorFaults uint64
	PageFaults  uint64
}

type ProcTime struct {
	StartTime uint64
	User      uint64
	Sys       uint64
	Total     uint64
}

type ProcArgs struct {
	List []string
}

type ProcExe struct {
	Name string
	Cwd  string
	Root string
}

type DiskList struct {
	List map[string]DiskIo
}

type DiskIo struct {
	ReadOps     uint64
	ReadBytes   uint64
	ReadTimeMs  uint64
	WriteOps    uint64
	WriteBytes  uint64
	WriteTimeMs uint64
	IoTimeMs    uint64
}

func (self DiskIo) Delta(other DiskIo) DiskIo {
	return DiskIo{
		ReadOps:     self.ReadOps - other.ReadOps,
		ReadBytes:   self.ReadBytes - other.ReadBytes,
		ReadTimeMs:  self.ReadTimeMs - other.ReadTimeMs,
		WriteOps:    self.WriteOps - other.WriteOps,
		WriteBytes:  self.WriteBytes - other.WriteBytes,
		WriteTimeMs: self.WriteTimeMs - other.WriteTimeMs,
		IoTimeMs:    self.IoTimeMs - other.IoTimeMs,
	}
}

type SystemInfo struct {
	Sysname    string
	Nodename   string
	Release    string
	Version    string
	Machine    string
	Domainname string
}

type SystemDistribution struct {
	Description string
}
