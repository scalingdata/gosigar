package sigar_test

import (
	"os"
	"fmt"
	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	sigar "github.com/scalingdata/gosigar"
)

var _ = Describe("SigarWindows", func() {
	Describe("Memory", func() {
		It("gets the total memory", func() {
			mem := sigar.Mem{}
			err := mem.Get()

			Ω(err).ShouldNot(HaveOccurred())
			Ω(mem.Total).Should(BeNumerically(">", 0))
		})
	})

	Describe("Memory", func() {
		It("gets the total swap", func() {
			swap := sigar.Swap{}
			err := swap.Get()

			Ω(err).ShouldNot(HaveOccurred())
			Ω(swap.Total).Should(BeNumerically(">", 0))
		})
	})

	Describe("FileSystemList", func() {
		It("gets volumes", func() {
			fsList := sigar.FileSystemList{}
			err := fsList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(fsList.List)).Should(BeNumerically(">", 0))
		})
	})

	Describe("Disk", func() {
		It("gets the total disk space", func() {
			usage := sigar.FileSystemUsage{}
			err := usage.Get(os.TempDir())

			Ω(err).ShouldNot(HaveOccurred())
			Ω(usage.Total).Should(BeNumerically(">", 0))
		})
	})

	Describe("DiskList", func() {
		It("gets the list of attached disks", func() {
			diskList := sigar.DiskList{}
			err := diskList.Get()

			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(diskList.List)).Should(BeNumerically(">", 0))
		})
	})

	Describe("Cpu", func() {
		It("gets CPU stats", func() {
			cpu := sigar.Cpu{}
			err := cpu.Get()

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cpu.Total()).Should(BeNumerically(">", 0))
		})
	})

	Describe("CpuList", func() {
		It("gets CPU stats", func() {
			cpuList := sigar.CpuList{}
			err := cpuList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cpuList.List[0].Total()).Should(BeNumerically(">", 0))
		})
	})

	Describe("NetIface", func() {
		It("gets interface stats", func() {
			netList := sigar.NetIfaceList{}
			err := netList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("IfaceList: %v\n", netList)
		})
	})

	Describe("NetConnList", func() {
		It("gets TCP IPv4 conns", func() {
			connList := sigar.NetTcpConnList{}
			err := connList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("Connections: %v\n", connList)
		})
		It("gets UDP IPv4 conns", func() {
			connList := sigar.NetUdpConnList{}
			err := connList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("Connections: %v\n", connList)
		})
		It("gets TCP IPv6 conns", func() {
			connList := sigar.NetTcpV6ConnList{}
			err := connList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("Connections: %v\n", connList)
		})
		It("gets UDP IPv6 conns", func() {
			connList := sigar.NetUdpV6ConnList{}
			err := connList.Get()
			Ω(err).ShouldNot(HaveOccurred())
			fmt.Printf("Connections: %v\n", connList)
		})
	})

	Describe("SystemInfo", func() {
		It("gets system info", func() {
			si := sigar.SystemInfo{}
			err := si.Get()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(si.Sysname).Should(Equal("Windows"))
		})
	})

	Describe("SystemDistribution", func() {
		It("gets system distribution", func() {
			sd := sigar.SystemDistribution{}
			err := sd.Get()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sd.Description).Should(Equal("Windows"))
		})
	})
})
