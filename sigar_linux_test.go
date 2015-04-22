package sigar_test

import (
	"io/ioutil"
	"time"

	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	sigar "github.com/scalingdata/gosigar"
)

var _ = Describe("sigarLinux", func() {
	var procd string

	BeforeEach(func() {
		var err error
		procd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())
		sigar.Procd = procd
	})

	AfterEach(func() {
		sigar.Procd = "/proc"
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
					Guest:  uint64(0),
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
					Guest:  uint64(0),
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
					Guest:  uint64(8),
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
					Guest:  uint64(80),
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
})
