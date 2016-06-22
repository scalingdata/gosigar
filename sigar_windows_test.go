package sigar_test

import (
	"fmt"
	"os"

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

	Describe("ProcessList", func() {
		It("gets the list of processes", func() {
			pl := sigar.ProcList{}
			err := pl.Get()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(pl.List)).Should(BeNumerically(">", 0))
		})
	})

	Describe("ProcessState", func() {
		It("gets the process detail", func() {
			pl := sigar.ProcList{}
			err := pl.Get()	
			Ω(err).ShouldNot(HaveOccurred())
			for _, pid := range pl.List {
				details := sigar.ProcState{}
				err  := details.Get(pid)
				fmt.Printf("Pid: %v Name: %v Err: %v\n", pid, details.Name, err)
			}
		})
	})

	Describe("ProcessTime", func() {
		It("gets the process detail", func() {
			pl := sigar.ProcList{}
			err := pl.Get()	
			Ω(err).ShouldNot(HaveOccurred())
			for _, pid := range pl.List {
				details := sigar.ProcTime{}
				err  := details.Get(pid)
				fmt.Printf("Pid: %v Start: %v Process: %v Sys: %v Err: %v\n", pid, details.StartTime, details.User, details.Sys, err)
			}
		})
	})

	Describe("ProcessMem", func() {
		It("gets the process detail", func() {
			pl := sigar.ProcList{}
			err := pl.Get()	
			Ω(err).ShouldNot(HaveOccurred())
			for _, pid := range pl.List {
				details := sigar.ProcMem{}
				err  := details.Get(pid)
				fmt.Printf("Pid: %v Resident: %v Size: %v Faults: %v Err: %v\n", pid, details.Resident, details.Size, details.PageFaults, err)
			}
		})
	})

	Describe("ProcessExe", func() {
		It("gets the process detail", func() {
			pl := sigar.ProcList{}
			err := pl.Get()	
			Ω(err).ShouldNot(HaveOccurred())
			for _, pid := range pl.List {
				details := sigar.ProcExe{}
				err  := details.Get(pid)
				fmt.Printf("Pid: %v Exe:%v Err: %v\n", pid, details.Name, err)
			}
		})
	})		
})
