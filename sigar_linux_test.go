package sigar_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	sigar "github.com/scalingdata/gosigar"
)

var _ = Describe("sigarLinux", func() {
	var procd string
	var sysd string

	BeforeEach(func() {
		var err error
		procd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())
		sysd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())

		sigar.Procd = procd
		sigar.Sysd = sysd
	})

	AfterEach(func() {
		sigar.Procd = "/proc"
		sigar.Sysd = "/sys"
	})

	Describe("CPU", func() {
		var (
			statFile string
			cpu      sigar.Cpu
		)

		BeforeEach(func() {
			statFile = procd + "/stat"
			cpu = sigar.Cpu{}
		})

		Describe("Get", func() {
			It("gets CPU usage", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.User).To(Equal(uint64(25)))
			})

			It("ignores empty lines", func() {
				statContents := []byte("cpu ")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.User).To(Equal(uint64(0)))
			})
			It("handles 2.6 CPU with guest stats", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7 8")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.Guest).To(Equal(uint64(8)))
			})
			It("handles 2.4 CPU without guest stats", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.Guest).To(Equal(uint64(0)))
			})
		})

		Describe("CollectCpuStats", func() {
			It("collects CPU usage over time (2.4)", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				concreteSigar := &sigar.ConcreteSigar{}
				cpuUsages, stop := concreteSigar.CollectCpuStats(500 * time.Millisecond)

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(25),
					Nice:    uint64(1),
					Sys:     uint64(2),
					Idle:    uint64(3),
					Wait:    uint64(4),
					Irq:     uint64(5),
					SoftIrq: uint64(6),
					Stolen:  uint64(7),
					Guest:   uint64(0),
				}))

				statContents = []byte("cpu 30 3 7 10 25 55 36 65")
				err = ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(5),
					Nice:    uint64(2),
					Sys:     uint64(5),
					Idle:    uint64(7),
					Wait:    uint64(21),
					Irq:     uint64(50),
					SoftIrq: uint64(30),
					Stolen:  uint64(58),
					Guest:   uint64(0),
				}))

				stop <- struct{}{}
			})
			It("collects CPU usage over time (2.6)", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7 8")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				concreteSigar := &sigar.ConcreteSigar{}
				cpuUsages, stop := concreteSigar.CollectCpuStats(500 * time.Millisecond)

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(25),
					Nice:    uint64(1),
					Sys:     uint64(2),
					Idle:    uint64(3),
					Wait:    uint64(4),
					Irq:     uint64(5),
					SoftIrq: uint64(6),
					Stolen:  uint64(7),
					Guest:   uint64(8),
				}))

				statContents = []byte("cpu 30 3 7 10 25 55 36 65 88")
				err = ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(5),
					Nice:    uint64(2),
					Sys:     uint64(5),
					Idle:    uint64(7),
					Wait:    uint64(21),
					Irq:     uint64(50),
					SoftIrq: uint64(30),
					Stolen:  uint64(58),
					Guest:   uint64(80),
				}))

				stop <- struct{}{}
			})
		})
	})

	Describe("Mem", func() {
		var meminfoFile string
		BeforeEach(func() {
			meminfoFile = procd + "/meminfo"

			meminfoContents := `
MemTotal:         374256 kB
MemFree:          274460 kB
Buffers:            9764 kB
Cached:            38648 kB
SwapCached:            0 kB
Active:            33772 kB
Inactive:          31184 kB
Active(anon):      16572 kB
Inactive(anon):      552 kB
Active(file):      17200 kB
Inactive(file):    30632 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:        786428 kB
SwapFree:         786428 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         16564 kB
Mapped:             6612 kB
Shmem:               584 kB
Slab:              19092 kB
SReclaimable:       9128 kB
SUnreclaim:         9964 kB
KernelStack:         672 kB
PageTables:         1864 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      973556 kB
Committed_AS:      55880 kB
VmallocTotal:   34359738367 kB
VmallocUsed:       21428 kB
VmallocChunk:   34359713596 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       59328 kB
DirectMap2M:      333824 kB
`
			err := ioutil.WriteFile(meminfoFile, []byte(meminfoContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns correct memory info", func() {
			mem := sigar.Mem{}
			err := mem.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(mem.Total).To(BeNumerically("==", 374256*1024))
			Expect(mem.Free).To(BeNumerically("==", 274460*1024))
		})
	})

	Describe("Swap", func() {
		var meminfoFile string
		BeforeEach(func() {
			meminfoFile = procd + "/meminfo"

			meminfoContents := `
MemTotal:         374256 kB
MemFree:          274460 kB
Buffers:            9764 kB
Cached:            38648 kB
SwapCached:            0 kB
Active:            33772 kB
Inactive:          31184 kB
Active(anon):      16572 kB
Inactive(anon):      552 kB
Active(file):      17200 kB
Inactive(file):    30632 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:        786428 kB
SwapFree:         786428 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         16564 kB
Mapped:             6612 kB
Shmem:               584 kB
Slab:              19092 kB
SReclaimable:       9128 kB
SUnreclaim:         9964 kB
KernelStack:         672 kB
PageTables:         1864 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      973556 kB
Committed_AS:      55880 kB
VmallocTotal:   34359738367 kB
VmallocUsed:       21428 kB
VmallocChunk:   34359713596 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       59328 kB
DirectMap2M:      333824 kB
`
			err := ioutil.WriteFile(meminfoFile, []byte(meminfoContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns correct memory info", func() {
			swap := sigar.Swap{}
			err := swap.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(swap.Total).To(BeNumerically("==", 786428*1024))
			Expect(swap.Free).To(BeNumerically("==", 786428*1024))
		})
	})

	Describe("DiskIO", func() {
		var diskstatFile string
		var partitionsFile string
		BeforeEach(func() {
			partitionsFile = procd + "/partitions"
			partitionsContents := `
major minor  #blocks  name

   8        0   41943040 sda
   8        1     512000 sda1
   8        2   41430016 sda2
 253        0   40476672 dm-0
 253        1     950272 dm-1
`
			diskstatFile = procd + "/diskstats"
			diskstatContents := `
   1       0 ram0 0 0 0 0 0 0 0 0 0 0 0
   1       1 ram1 0 0 0 0 0 0 0 0 0 0 0
   1       2 ram2 0 0 0 0 0 0 0 0 0 0 0
   1       3 ram3 0 0 0 0 0 0 0 0 0 0 0
   1       4 ram4 0 0 0 0 0 0 0 0 0 0 0
   1       5 ram5 0 0 0 0 0 0 0 0 0 0 0
   1       6 ram6 0 0 0 0 0 0 0 0 0 0 0
   1       7 ram7 0 0 0 0 0 0 0 0 0 0 0
   1       8 ram8 0 0 0 0 0 0 0 0 0 0 0
   1       9 ram9 0 0 0 0 0 0 0 0 0 0 0
   1      10 ram10 0 0 0 0 0 0 0 0 0 0 0
   1      11 ram11 0 0 0 0 0 0 0 0 0 0 0
   1      12 ram12 0 0 0 0 0 0 0 0 0 0 0
   1      13 ram13 0 0 0 0 0 0 0 0 0 0 0
   1      14 ram14 0 0 0 0 0 0 0 0 0 0 0
   1      15 ram15 0 0 0 0 0 0 0 0 0 0 0
   7       0 loop0 0 0 0 0 0 0 0 0 0 0 0
   7       1 loop1 0 0 0 0 0 0 0 0 0 0 0
   7       2 loop2 0 0 0 0 0 0 0 0 0 0 0
   7       3 loop3 0 0 0 0 0 0 0 0 0 0 0
   7       4 loop4 0 0 0 0 0 0 0 0 0 0 0
   7       5 loop5 0 0 0 0 0 0 0 0 0 0 0
   7       6 loop6 0 0 0 0 0 0 0 0 0 0 0
   7       7 loop7 0 0 0 0 0 0 0 0 0 0 0
   8       0 sda 7089 1514 243714 6329 86342 820181 7159936 1397746 0 50380 1404142
   8       1 sda1 560 243 12192 133 64 5182 10504 955 0 218 1085
   8       2 sda2 6375 1271 230290 6161 77130 814999 7149432 1395024 0 48563 1401264
 253       0 dm-0 7255 0 226842 8638 895753 0 7149432 40185506 0 50361 40194137
 253       1 dm-1 307 0 2456 116 0 0 0 0 0 116 116
`
			err := ioutil.WriteFile(partitionsFile, []byte(partitionsContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			err = ioutil.WriteFile(diskstatFile, []byte(diskstatContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("skips non-block devices", func() {
			ioStat := sigar.DiskList{}
			err := ioStat.Get()
			Expect(err).ToNot(HaveOccurred())

			_, ok := ioStat.List["ram0"]
			Expect(ok).To(Equal(false))
			_, ok = ioStat.List["loop0"]
			Expect(ok).To(Equal(false))
		})

		It("returns correct disk IO info", func() {
			ioStat := sigar.DiskList{}
			err := ioStat.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(ioStat.List["sda"].ReadOps).To(Equal(uint64(7089)))
			Expect(ioStat.List["sda"].ReadBytes).To(Equal(uint64(243714 * 512)))
			Expect(ioStat.List["sda"].ReadTimeMs).To(Equal(uint64(6329)))
			Expect(ioStat.List["sda"].WriteOps).To(Equal(uint64(86342)))
			Expect(ioStat.List["sda"].WriteBytes).To(Equal(uint64(7159936 * 512)))
			Expect(ioStat.List["sda"].WriteTimeMs).To(Equal(uint64(1397746)))
			Expect(ioStat.List["sda"].IoTimeMs).To(Equal(uint64(50380)))
		})
	})

	Describe("NetIface", func() {
		var netDevFile string
		BeforeEach(func() {
			netDevFile = procd + "/net/dev"
			netDevContents := `
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:   32814     457    7    4    3     8          0         2   32814      457    0    8    0    78       1          8
  eth0: 14071306   63882   25  100    4  1000         10         0  3729191   37066    7    0   11    20       0          9
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netDevFile, []byte(netDevContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			netEth0MtuFile := sysd + "/class/net/eth0/mtu"
			netEth0AddrFile := sysd + "/class/net/eth0/address"
			netEth0CarrierFile := sysd + "/class/net/eth0/carrier"
			mtu := "1500"
			address := "08:00:27:6b:1c:dd"
			linkStat := "1"

			err = os.MkdirAll(sysd+"/class/net/eth0/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netEth0MtuFile, []byte(mtu), 0444)
			err = ioutil.WriteFile(netEth0AddrFile, []byte(address), 0444)
			err = ioutil.WriteFile(netEth0CarrierFile, []byte(linkStat), 0444)
			Expect(err).ToNot(HaveOccurred())

		})

		It("parses network interface stats", func() {
			netStat := sigar.NetIfaceList{}
			err := netStat.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(netStat.List[0].Name).To(Equal("lo"))
			Expect(netStat.List[0].SendBytes).To(Equal(uint64(32814)))
			Expect(netStat.List[0].RecvBytes).To(Equal(uint64(32814)))
			Expect(netStat.List[0].SendPackets).To(Equal(uint64(457)))
			Expect(netStat.List[0].RecvPackets).To(Equal(uint64(457)))
			Expect(netStat.List[0].SendCompressed).To(Equal(uint64(8)))
			Expect(netStat.List[0].RecvCompressed).To(Equal(uint64(0)))
			Expect(netStat.List[0].RecvMulticast).To(Equal(uint64(2)))
			Expect(netStat.List[0].SendErrors).To(Equal(uint64(0)))
			Expect(netStat.List[0].RecvErrors).To(Equal(uint64(7)))
			Expect(netStat.List[0].SendDropped).To(Equal(uint64(8)))
			Expect(netStat.List[0].RecvDropped).To(Equal(uint64(4)))
			Expect(netStat.List[0].SendFifoErrors).To(Equal(uint64(0)))
			Expect(netStat.List[0].RecvFifoErrors).To(Equal(uint64(3)))
			Expect(netStat.List[0].RecvFramingErrors).To(Equal(uint64(8)))
			Expect(netStat.List[0].SendCarrier).To(Equal(uint64(1)))
			Expect(netStat.List[0].SendCollisions).To(Equal(uint64(78)))

			Expect(netStat.List[1].Name).To(Equal("eth0"))
			Expect(netStat.List[1].SendBytes).To(Equal(uint64(3729191)))
			Expect(netStat.List[1].RecvBytes).To(Equal(uint64(14071306)))
			Expect(netStat.List[1].SendPackets).To(Equal(uint64(37066)))
			Expect(netStat.List[1].RecvPackets).To(Equal(uint64(63882)))
			Expect(netStat.List[1].SendCompressed).To(Equal(uint64(9)))
			Expect(netStat.List[1].RecvCompressed).To(Equal(uint64(10)))
			Expect(netStat.List[1].RecvMulticast).To(Equal(uint64(0)))
			Expect(netStat.List[1].SendErrors).To(Equal(uint64(7)))
			Expect(netStat.List[1].RecvErrors).To(Equal(uint64(25)))
			Expect(netStat.List[1].SendDropped).To(Equal(uint64(0)))
			Expect(netStat.List[1].RecvDropped).To(Equal(uint64(100)))
			Expect(netStat.List[1].SendFifoErrors).To(Equal(uint64(11)))
			Expect(netStat.List[1].RecvFifoErrors).To(Equal(uint64(4)))
			Expect(netStat.List[1].RecvFramingErrors).To(Equal(uint64(1000)))
			Expect(netStat.List[1].SendCarrier).To(Equal(uint64(0)))
			Expect(netStat.List[1].SendCollisions).To(Equal(uint64(20)))
		})

		It("parses network interface info when available", func() {
			netStat := sigar.NetIfaceList{}
			err := netStat.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(netStat.List[0].MTU).To(Equal(uint64(0)))
			Expect(netStat.List[0].Mac).To(Equal(""))
			Expect(netStat.List[0].LinkStatus).To(Equal(""))

			Expect(netStat.List[1].MTU).To(Equal(uint64(1500)))
			Expect(netStat.List[1].Mac).To(Equal("08:00:27:6b:1c:dd"))
			Expect(netStat.List[1].LinkStatus).To(Equal("UP"))
		})
	})

	Describe("SystemInfo", func() {
		var (
			si sigar.SystemInfo
		)

		BeforeEach(func() {
			si = sigar.SystemInfo{}
		})

		Describe("Get", func() {
			It("gets System Information", func() {
				err := si.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(si.Sysname).To(Equal("Linux"))
			})
		})
	})

	Describe("SystemDistribution", func() {
		var (
			sd sigar.SystemDistribution
		)

		BeforeEach(func() {
			sd = sigar.SystemDistribution{}
		})

		Describe("Get", func() {
			It("gets System Distribution", func() {
				err := sd.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(sd.Description).ToNot(Equal(""))
			})
		})
	})

	Describe("Process", func() {
		It("GetsProcessList", func() {
			err := os.MkdirAll(procd+"/stat", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())

			procList := &sigar.ProcList{}
			err = procList.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(len(procList.List)).To(Equal(1))
			Expect(procList.List[0]).To(Equal(10))
		})

		It("GetsArgs", func() {
			cmdLineFile := procd + "/10/cmdline"
			cmdLine := "/sbin/min\x00getty\x00/dev/tty3\x00"
			err := os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(cmdLineFile, []byte(cmdLine), 0444)
			Expect(err).ToNot(HaveOccurred())

			procArgs := &sigar.ProcArgs{}
			err = procArgs.Get(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(procArgs.List).To(Equal([]string{"/sbin/min", "getty", "/dev/tty3"}))
		})

		It("GetsProcessState", func() {
			statFile := procd + "/10/stat"
			statLine := "10 (watchdog/1) S 2 0 0 11 -1 2216722752 0 0 0 0 0 142 0 0 -100 0 1 0 4 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 1 99 1 0 0 0"
			err := os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(statFile, []byte(statLine), 0444)
			Expect(err).ToNot(HaveOccurred())

			procState := &sigar.ProcState{}
			err = procState.Get(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(procState.Name).To(Equal("watchdog/1"))
			Expect(procState.State).To(Equal(sigar.RunState(sigar.RunStateSleep)))
			Expect(procState.Ppid).To(Equal(int(2)))
			Expect(procState.Tty).To(Equal(int(11)))
			Expect(procState.Priority).To(Equal(int(-100)))
			Expect(procState.Nice).To(Equal(int(0)))
			Expect(procState.Processor).To(Equal(int(1)))
		})

		It("GetsProcessTime", func() {
			statFile := procd + "/10/stat"
			statLine := "10 (watchdog/1) S 2 0 0 11 -1 2216722752 0 0 0 0 100 142 0 0 -100 0 1 0 40000 240 160 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 1 99 1 0 0 0"
			err := os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(statFile, []byte(statLine), 0444)
			Expect(err).ToNot(HaveOccurred())
			// Process start time is relative to system start time, we need to reset the system start time
			bTimeFile := procd + "/stat"
			bTimeContent := "btime: 0"
			err = ioutil.WriteFile(bTimeFile, []byte(bTimeContent), 0444)
			Expect(err).ToNot(HaveOccurred())
			sigar.LoadStartTime()

			procTime := &sigar.ProcTime{}
			err = procTime.Get(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(procTime.User).To(Equal(uint64(1000)))
			Expect(procTime.Sys).To(Equal(uint64(1420)))
			Expect(procTime.Total).To(Equal(uint64(2420)))
			Expect(procTime.StartTime).To(Equal(uint64(400000)))
		})

		It("GetsProcessMemory", func() {
			statFile := procd + "/10/stat"
			statLine := "10 (watchdog/1) S 2 0 0 11 -1 2216722752 0 64 128 256 0 142 0 0 -100 0 1 120 4 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 1 99 1 0 0 0"
			statmFile := procd + "/10/statm"
			statmLine := "63831 465 293 89 0 56957 0"
			err := os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(statFile, []byte(statLine), 0444)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(statmFile, []byte(statmLine), 0444)
			Expect(err).ToNot(HaveOccurred())

			procMem := &sigar.ProcMem{}
			err = procMem.Get(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(procMem.Size).To(Equal(uint64(261451776)))
			Expect(procMem.Resident).To(Equal(uint64(1904640)))
			Expect(procMem.Share).To(Equal(uint64(1200128)))
			Expect(procMem.MinorFaults).To(Equal(uint64(64)))
			Expect(procMem.MajorFaults).To(Equal(uint64(256)))
			Expect(procMem.PageFaults).To(Equal(uint64(320)))
		})

		It("GetsProcessIo", func() {
			ioFile := procd + "/10/io"
			ioFileContents := `
rchar: 5811
wchar: 949188
syscr: 4
syscw: 4053
read_bytes: 12288
write_bytes: 5365760
cancelled_write_bytes: 0
`
			err := os.MkdirAll(procd+"/10/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(ioFile, []byte(ioFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			procIo := &sigar.ProcIo{}
			err = procIo.Get(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(procIo.ReadOps).To(Equal(uint64(4)))
			Expect(procIo.WriteOps).To(Equal(uint64(4053)))
			Expect(procIo.ReadBytes).To(Equal(uint64(12288)))
			Expect(procIo.WriteBytes).To(Equal(uint64(5365760)))
		})
	})
})
