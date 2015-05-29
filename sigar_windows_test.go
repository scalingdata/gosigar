package sigar_test

import (
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
})
