package sigar_test

import (
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	sigar "github.com/scalingdata/gosigar"
)

var _ = Describe("sigarLinux", func() {
	var procd string
	var sysd string
	var etcd string

	BeforeEach(func() {
		var err error
		procd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())
		sysd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())
		etcd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())

		sigar.Procd = procd
		sigar.Sysd = sysd
		sigar.Etcd = etcd
	})

	AfterEach(func() {
		sigar.Procd = "/proc"
		sigar.Sysd = "/sys"
		sigar.Etcd = "/etc"
	})

	It("Parses integers correctly", func() {
		Expect(sigar.ReadUint("123")).To(Equal(uint64(123)))
		Expect(sigar.ReadUint("123\n456")).To(Equal(uint64(0)))
		Expect(sigar.ReadUint("-1")).To(Equal(uint64(0)))
		Expect(sigar.ReadUint("abc")).To(Equal(uint64(0)))
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

	Describe("NetProtoV4", func() {
		BeforeEach(func() {
			netSnmpContents := ` 
Ip: Forwarding DefaultTTL InReceives InHdrErrors InAddrErrors ForwDatagrams InUnknownProtos InDiscards InDelivers OutRequests OutDiscards OutNoRoutes ReasmTimeout ReasmReqds ReasmOKs ReasmFails FragOKs FragFails FragCreates
Ip: 2 64 1055501 33 56 1 44 2 1055445 503315 55 3 0 0 0 0 0 0 0
Icmp: InMsgs InErrors InDestUnreachs InTimeExcds InParmProbs InSrcQuenchs InRedirects InEchos InEchoReps InTimestamps InTimestampReps InAddrMasks InAddrMaskReps OutMsgs OutErrors OutDestUnreachs OutTimeExcds OutParmProbs OutSrcQuenchs OutRedirects OutEchos OutEchoReps OutTimestamps OutTimestampReps OutAddrMasks OutAddrMaskReps
Icmp: 83 1 12 0 0 0 0 30 41 0 0 0 0 71 2 33 0 0 0 0 41 30 0 0 0 0
IcmpMsg: InType0 InType3 InType8 OutType0 OutType8
IcmpMsg: 41 12 30 30 41
Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens PassiveOpens AttemptFails EstabResets CurrEstab InSegs OutSegs RetransSegs InErrs OutRsts
Tcp: 1 200 120000 -1 1189 386 385 66 2 1053372 501082 24 99 526
Udp: InDatagrams NoPorts InErrors OutDatagrams RcvbufErrors SndbufErrors
Udp: 1997 77 1 2145 99 88
UdpLite: InDatagrams NoPorts InErrors OutDatagrams RcvbufErrors SndbufErrors
UdpLite: 0 0 0 0 0 0`
			netSnmpFile := procd + "/net/snmp"
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netSnmpFile, []byte(netSnmpContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("parses network protocol stats", func() {
			netStat := sigar.NetProtoV4Stats{}
			err := netStat.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(netStat.IP.InReceives).To(Equal(uint64(1055501)))
			Expect(netStat.IP.InHdrErrors).To(Equal(uint64(33)))
			Expect(netStat.IP.InAddrErrors).To(Equal(uint64(56)))
			Expect(netStat.IP.ForwDatagrams).To(Equal(uint64(1)))
			Expect(netStat.IP.InDelivers).To(Equal(uint64(1055445)))
			Expect(netStat.IP.InDiscards).To(Equal(uint64(2)))
			Expect(netStat.IP.InUnknownProtos).To(Equal(uint64(44)))
			Expect(netStat.IP.OutRequests).To(Equal(uint64(503315)))
			Expect(netStat.IP.OutDiscards).To(Equal(uint64(55)))
			Expect(netStat.IP.OutNoRoutes).To(Equal(uint64(3)))

			Expect(netStat.ICMP.InMsgs).To(Equal(uint64(83)))
			Expect(netStat.ICMP.InErrors).To(Equal(uint64(1)))
			Expect(netStat.ICMP.InDestUnreachs).To(Equal(uint64(12)))
			Expect(netStat.ICMP.OutMsgs).To(Equal(uint64(71)))
			Expect(netStat.ICMP.OutErrors).To(Equal(uint64(2)))
			Expect(netStat.ICMP.OutDestUnreachs).To(Equal(uint64(33)))

			Expect(netStat.TCP.ActiveOpens).To(Equal(uint64(1189)))
			Expect(netStat.TCP.PassiveOpens).To(Equal(uint64(386)))
			Expect(netStat.TCP.AttemptFails).To(Equal(uint64(385)))
			Expect(netStat.TCP.EstabResets).To(Equal(uint64(66)))
			Expect(netStat.TCP.CurrEstab).To(Equal(uint64(2)))
			Expect(netStat.TCP.InSegs).To(Equal(uint64(1053372)))
			Expect(netStat.TCP.OutSegs).To(Equal(uint64(501082)))
			Expect(netStat.TCP.RetransSegs).To(Equal(uint64(24)))
			Expect(netStat.TCP.InErrs).To(Equal(uint64(99)))
			Expect(netStat.TCP.OutRsts).To(Equal(uint64(526)))

			Expect(netStat.UDP.InDatagrams).To(Equal(uint64(1997)))
			Expect(netStat.UDP.OutDatagrams).To(Equal(uint64(2145)))
			Expect(netStat.UDP.InErrors).To(Equal(uint64(1)))
			Expect(netStat.UDP.NoPorts).To(Equal(uint64(77)))
			Expect(netStat.UDP.RcvbufErrors).To(Equal(uint64(99)))
			Expect(netStat.UDP.SndbufErrors).To(Equal(uint64(88)))
		})
	})

	Describe("NetProtoV6", func() {
		BeforeEach(func() {
			netSnmpContents := `
Ip6InReceives                   	1147
Ip6InHdrErrors                  	0
Ip6InTooBigErrors               	0
Ip6InNoRoutes                   	0
Ip6InAddrErrors                 	10
Ip6InUnknownProtos              	0
Ip6InTruncatedPkts              	0
Ip6InDiscards                   	13
Ip6InDelivers                   	1146
Ip6OutForwDatagrams             	11
Ip6OutRequests                  	1224
Ip6OutDiscards                  	0
Ip6OutNoRoutes                  	21
Ip6ReasmTimeout                 	0
Ip6ReasmReqds                   	0
Ip6ReasmOKs                     	0
Ip6ReasmFails                   	0
Ip6FragOKs                      	0
Ip6FragFails                    	0
Ip6FragCreates                  	0
Ip6InMcastPkts                  	0
Ip6OutMcastPkts                 	106
Ip6InOctets                     	119080
Ip6OutOctets                    	124428
Ip6InMcastOctets                	0
Ip6OutMcastOctets               	7912
Ip6InBcastOctets                	0
Ip6OutBcastOctets               	0
Icmp6InMsgs                     	1140
Icmp6InErrors                   	50
Icmp6OutMsgs                    	1217
Icmp6InDestUnreachs             	51
Icmp6InPktTooBigs               	0
Icmp6InTimeExcds                	0
Icmp6InParmProblems             	0
Icmp6InEchos                    	570
Icmp6InEchoReplies              	570
Icmp6InGroupMembQueries         	0
Icmp6InGroupMembResponses       	0
Icmp6InGroupMembReductions      	0
Icmp6InRouterSolicits           	0
Icmp6InRouterAdvertisements     	0
Icmp6InNeighborSolicits         	0
Icmp6InNeighborAdvertisements   	0
Icmp6InRedirects                	0
Icmp6InMLDv2Reports             	0
Icmp6OutDestUnreachs            	52
Icmp6OutPktTooBigs              	0
Icmp6OutTimeExcds               	0
Icmp6OutParmProblems            	0
Icmp6OutEchos                   	570
Icmp6OutEchoReplies             	570
Icmp6OutGroupMembQueries        	0
Icmp6OutGroupMembResponses      	0
Icmp6OutGroupMembReductions     	0
Icmp6OutRouterSolicits          	36
Icmp6OutRouterAdvertisements    	0
Icmp6OutNeighborSolicits        	12
Icmp6OutNeighborAdvertisements  	0
Icmp6OutRedirects               	0
Icmp6OutMLDv2Reports            	29
Icmp6InType128                  	570
Icmp6InType129                  	570
Icmp6OutType128                 	570
Icmp6OutType129                 	570
Icmp6OutType133                 	36
Icmp6OutType135                 	12
Icmp6OutType143                 	29
Udp6InDatagrams                 	700
Udp6NoPorts                     	750
Udp6InErrors                    	751
Udp6OutDatagrams                	752
UdpLite6InDatagrams             	0
UdpLite6NoPorts                 	0
UdpLite6InErrors                	0
UdpLite6OutDatagrams            	0
`
			netSnmpFile := procd + "/net/snmp6"
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netSnmpFile, []byte(netSnmpContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("parses network protocol stats", func() {
			netStat := sigar.NetProtoV6Stats{}
			err := netStat.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(netStat.IP.InReceives).To(Equal(uint64(1147)))
			Expect(netStat.IP.InAddrErrors).To(Equal(uint64(10)))
			Expect(netStat.IP.ForwDatagrams).To(Equal(uint64(11)))
			Expect(netStat.IP.InDelivers).To(Equal(uint64(1146)))
			Expect(netStat.IP.InDiscards).To(Equal(uint64(13)))
			Expect(netStat.IP.OutRequests).To(Equal(uint64(1224)))

			Expect(netStat.ICMP.InMsgs).To(Equal(uint64(1140)))
			Expect(netStat.ICMP.InErrors).To(Equal(uint64(50)))
			Expect(netStat.ICMP.InDestUnreachs).To(Equal(uint64(51)))
			Expect(netStat.ICMP.OutMsgs).To(Equal(uint64(1217)))
			Expect(netStat.ICMP.OutErrors).To(Equal(uint64(0))) // Not reported by snmp6
			Expect(netStat.ICMP.OutDestUnreachs).To(Equal(uint64(52)))

			Expect(netStat.UDP.InDatagrams).To(Equal(uint64(700)))
			Expect(netStat.UDP.OutDatagrams).To(Equal(uint64(752)))
			Expect(netStat.UDP.InErrors).To(Equal(uint64(751)))
			Expect(netStat.UDP.NoPorts).To(Equal(uint64(750)))
			Expect(netStat.UDP.RcvbufErrors).To(Equal(uint64(0))) // Not reported by snmp6
			Expect(netStat.UDP.SndbufErrors).To(Equal(uint64(0))) // Not reported by snmp6
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

			// Sample values. The newline is present on a real system
			mtu := "1500\n"
			address := "08:00:27:6b:1c:dd\n"
			linkStat := "1\n"

			err = os.MkdirAll(sysd+"/class/net/eth0/", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netEth0MtuFile, []byte(mtu), 0444)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(netEth0AddrFile, []byte(address), 0444)
			Expect(err).ToNot(HaveOccurred())
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
			Expect(netStat.List[0].LinkStatus).To(Equal("UNKNOWN"))

			Expect(netStat.List[1].MTU).To(Equal(uint64(1500)))
			Expect(netStat.List[1].Mac).To(Equal("08:00:27:6b:1c:dd"))
			Expect(netStat.List[1].LinkStatus).To(Equal("UP"))
		})
	})

	Describe("NetConn", func() {
		It("parses TCP IPv4", func() {
			connFile := procd + "/net/tcp"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
   0: 00000000:0016 00000000:0000 0A 00000000:00000123 00:00000000 00000000     0        0 12095 1 ffff880296063500 99 0 0 10 -1                     
   3: 0F02000A:0016 0202000A:E78C 01 00000490:00000000 02:00050277 00000000     0        0 95158 3 ffff880297be4e80 20 5 25 10 -1                    
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetTcpConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(2))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(22)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(0)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(123)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateListen))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoTcp))
			Expect(connList.List[0].String()).To(Equal("Listen tcp 0.0.0.0:22"))

			Expect(connList.List[1].LocalAddr).To(Equal(net.IP{10, 0, 2, 15}))
			Expect(connList.List[1].RemoteAddr).To(Equal(net.IP{10, 0, 2, 2}))
			Expect(connList.List[1].LocalPort).To(Equal(uint64(22)))
			Expect(connList.List[1].RemotePort).To(Equal(uint64(59276)))
			Expect(connList.List[1].SendQueue).To(Equal(uint64(490)))
			Expect(connList.List[1].RecvQueue).To(Equal(uint64(0)))
			Expect(connList.List[1].Status).To(Equal(sigar.ConnStateEstablished))
			Expect(connList.List[1].Proto).To(Equal(sigar.ConnProtoTcp))
			Expect(connList.List[1].String()).To(Equal("tcp 10.0.2.15:22 <-> 10.0.2.2:59276"))
		})

		It("parses UDP IPv4", func() {
			connFile := procd + "/net/udp"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
  19: 00000000:0044 00000000:0000 07 00000000:00000123 00:00000000 00000000     0        0 180235 2 ffff880297434080 0         
  62: 0F02000A:006F 00000000:0000 07 00000490:00000000 00:00000000 00000000     0        0 11060 2 ffff880296c92ac0 0 
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetUdpConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(2))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(68)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(0)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(123)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateClose))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoUdp))
			Expect(connList.List[0].String()).To(Equal("udp 0.0.0.0:68 <-> 0.0.0.0:0"))

			Expect(connList.List[1].LocalAddr).To(Equal(net.IP{10, 0, 2, 15}))
			Expect(connList.List[1].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[1].LocalPort).To(Equal(uint64(111)))
			Expect(connList.List[1].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[1].SendQueue).To(Equal(uint64(490)))
			Expect(connList.List[1].RecvQueue).To(Equal(uint64(0)))
			Expect(connList.List[1].Status).To(Equal(sigar.ConnStateClose))
			Expect(connList.List[1].Proto).To(Equal(sigar.ConnProtoUdp))
			Expect(connList.List[1].String()).To(Equal("udp 10.0.2.15:111 <-> 0.0.0.0:0"))
		})

		It("parses raw IPv4", func() {
			connFile := procd + "/net/raw"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
   1: 00000000:0001 00000000:0000 07 00000011:00000540 00:00000000 00000000     0        0 201340 2 ffff88029af250c0 0
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetRawConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(1))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(1)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(11)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(540)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateClose))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoRaw))
			Expect(connList.List[0].String()).To(Equal("raw 0.0.0.0:1 <-> 0.0.0.0:0"))
		})

		It("parses TCP IPv6", func() {
			connFile := procd + "/net/tcp6"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
   2: 00000000000000000000000001000000:9E32 00000000000000000000000001000000:0955 06 00000128:00000512 03:000015ED 00000000     0        0 0 3 ffff88029741cd00 99 0 0 2 -1
   3: 00000000000000000000000001000000:0955 00000000000000000000000001000000:9E32 06 00007890:00000111 03:000015ED 00000000     0        0 0 3 ffff88029741ce40 99 0 0 2 -1
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetTcpV6ConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(2))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(40498)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(2389)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(128)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(512)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateTimeWait))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoTcp))
			Expect(connList.List[0].String()).To(Equal("tcp ::1:40498 <-> ::1:2389"))

			Expect(connList.List[1].LocalAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[1].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[1].LocalPort).To(Equal(uint64(2389)))
			Expect(connList.List[1].RemotePort).To(Equal(uint64(40498)))
			Expect(connList.List[1].SendQueue).To(Equal(uint64(7890)))
			Expect(connList.List[1].RecvQueue).To(Equal(uint64(111)))
			Expect(connList.List[1].Status).To(Equal(sigar.ConnStateTimeWait))
			Expect(connList.List[1].Proto).To(Equal(sigar.ConnProtoTcp))
			Expect(connList.List[1].String()).To(Equal("tcp ::1:2389 <-> ::1:40498"))
		})

		It("parses UDP IPv6", func() {
			connFile := procd + "/net/udp6"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
  62: 00000000000000000000000000000000:006F 00000000000000000000000000000000:0000 07 00000128:00000512 00:00000000 00000000     0        0 11065 2 ffff880297b59c00 0
  88: 00000000000000000000000001000000:DC89 00000000000000000000000001000000:03E9 01 00007890:00000111 00:00000000 00000000     0        0 203648 2 ffff880296d07400 0
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetUdpV6ConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(2))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(111)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(128)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(512)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateClose))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoUdp))
			Expect(connList.List[0].String()).To(Equal("udp :::111 <-> :::0"))

			Expect(connList.List[1].LocalAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[1].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}))
			Expect(connList.List[1].LocalPort).To(Equal(uint64(56457)))
			Expect(connList.List[1].RemotePort).To(Equal(uint64(1001)))
			Expect(connList.List[1].SendQueue).To(Equal(uint64(7890)))
			Expect(connList.List[1].RecvQueue).To(Equal(uint64(111)))
			Expect(connList.List[1].Status).To(Equal(sigar.ConnStateEstablished))
			Expect(connList.List[1].Proto).To(Equal(sigar.ConnProtoUdp))
			Expect(connList.List[1].String()).To(Equal("udp ::1:56457 <-> ::1:1001"))
		})

		It("parses raw IPv6", func() {
			connFile := procd + "/net/raw6"
			connFileContents := `
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
  58: 00000000000000000000000000000000:003A 00000000000000000000000000000000:0000 07 00000000:00000000 00:00000000 00000000     0        0 201786 2 ffff88029a347000 0
`
			err := os.MkdirAll(procd+"/net", 0777)
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(connFile, []byte(connFileContents), 0444)
			Expect(err).ToNot(HaveOccurred())

			connList := sigar.NetRawV6ConnList{}
			err = connList.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(connList.List)).To(Equal(1))
			Expect(connList.List[0].LocalAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			Expect(connList.List[0].RemoteAddr).To(Equal(net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			Expect(connList.List[0].LocalPort).To(Equal(uint64(58)))
			Expect(connList.List[0].RemotePort).To(Equal(uint64(0)))
			Expect(connList.List[0].SendQueue).To(Equal(uint64(0)))
			Expect(connList.List[0].RecvQueue).To(Equal(uint64(0)))
			Expect(connList.List[0].Status).To(Equal(sigar.ConnStateClose))
			Expect(connList.List[0].Proto).To(Equal(sigar.ConnProtoRaw))
			Expect(connList.List[0].String()).To(Equal("raw :::58 <-> :::0"))
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
			sd                    sigar.SystemDistribution
			redhatReleaseFile     string
			redhatReleaseContents string
			lsbReleaseFile        string
			lsbReleaseContents    string
		)

		BeforeEach(func() {
			sd = sigar.SystemDistribution{}
			redhatReleaseFile = etcd + "/redhat-release"
			redhatReleaseContents = "CentOS release 6.8 (Final)\n"

			lsbReleaseFile = etcd + "/lsb-release"
			lsbReleaseContents = `DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=12.04
DISTRIB_CODENAME=precise
DISTRIB_DESCRIPTION="Ubuntu 12.04.5 LTS"
`

			// Always write lsb-release. The redhat-release file is only written by tests that use it
			err := ioutil.WriteFile(lsbReleaseFile, []byte(lsbReleaseContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(lsbReleaseFile)
			os.Remove(redhatReleaseFile)
		})

		Describe("Get", func() {
			It("gets System Distribution from redhat-release, if present", func() {
				err := ioutil.WriteFile(redhatReleaseFile, []byte(redhatReleaseContents), 0444)
				Expect(err).ToNot(HaveOccurred())

				err = sd.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(sd.Description).To(Equal(strings.TrimSpace(redhatReleaseContents)))
			})

			It("gets System Distribution from lsb-release", func() {
				err := sd.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(sd.Description).To(Equal("Ubuntu 12.04.5 LTS"))
			})

			It("gets System Distribution from lsb-release with blank distribution description", func() {
				err := os.Remove(lsbReleaseFile)
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(lsbReleaseFile, []byte("DISTRIB_DESCRIPTION="), 0444)
				Expect(err).ToNot(HaveOccurred())

				err = sd.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(sd.Description).To(Equal(""))
			})

			It("errors when lsb-release is not present", func() {
				err := os.Remove(lsbReleaseFile)
				Expect(err).ToNot(HaveOccurred())

				err = sd.Get()
				Expect(err).To(HaveOccurred())
				Expect(sd.Description).To(Equal(""))
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
			Expect(procState.Pid).To(Equal(10))
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
