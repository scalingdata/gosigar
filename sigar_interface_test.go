package sigar_test

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	. "github.com/scalingdata/gosigar"
)

var _ = Describe("Sigar", func() {
	var invalidPid = 666666

	It("cpu", func() {
		cpu := Cpu{}
		err := cpu.Get()
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("load average", func() {
		avg := LoadAverage{}
		err := avg.Get()

		if runtime.GOOS == "windows" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("uptime", func() {
		uptime := Uptime{}
		err := uptime.Get()
		if runtime.GOOS == "windows" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(uptime.Length).To(BeNumerically(">", float64(0)))
		}
	})

	It("mem", func() {
		mem := Mem{}
		err := mem.Get()
		Expect(err).ToNot(HaveOccurred())

		Expect(mem.Total).To(BeNumerically(">", 0))
		Expect(mem.Used + mem.Free).To(BeNumerically("<=", mem.Total))
	})

	It("swap", func() {
		swap := Swap{}
		err := swap.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(swap.Used + swap.Free).To(BeNumerically("<=", swap.Total))
	})

	It("cpu list", func() {
		cpulist := CpuList{}
		err := cpulist.Get()
		Expect(err).ToNot(HaveOccurred())

		nsigar := len(cpulist.List)
		numcpu := runtime.NumCPU()
		Expect(nsigar).To(Equal(numcpu))
	})

	It("file system list", func() {
		fslist := FileSystemList{}
		err := fslist.Get()
		Expect(err).ToNot(HaveOccurred())

		Expect(len(fslist.List)).To(BeNumerically(">", 0))
	})

	It("file system usage", func() {
		fsusage := FileSystemUsage{}
		err := fsusage.Get("/")
		Expect(err).ToNot(HaveOccurred())

		err = fsusage.Get("T O T A L L Y B O G U S")
		Expect(err).To(HaveOccurred())
	})

	It("net proto v4", func() {
		net := NetProtoV4Stats{}
		err := net.Get()
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("net proto v6", func() {
		net := NetProtoV6Stats{}
		err := net.Get()
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("net interface", func() {
		net := NetIfaceList{}
		err := net.Get()
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("net connection lists", func() {
		// Test connLists available on Windows/Linux
		connLists := []Getter{
			&NetTcpConnList{},
			&NetUdpConnList{},
			&NetTcpV6ConnList{},
			&NetUdpV6ConnList{},
		}
		for _, connList := range connLists {
			err := connList.Get()
			if runtime.GOOS == "darwin" {
				Expect(err).To(Equal(ErrNotImplemented))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		}

		// Test connLists available only on Linux
		connLists = []Getter{
			&NetRawConnList{},
			&NetRawV6ConnList{},
		}
		for _, connList := range connLists {
			err := connList.Get()
			if runtime.GOOS == "linux" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(Equal(ErrNotImplemented))
			}
		}
	})

	It("full process list", func() {
		processList := ProcessList{}
		err := processList.Get()
		if runtime.GOOS == "windows" {
			Expect(err).ToNot(HaveOccurred())
			Expect(len(processList.List)).To(BeNumerically(">", 0))
		} else {
			Expect(err).To(Equal(ErrNotImplemented))
		}
	})

	It("proc list", func() {
		pids := ProcList{}
		err := pids.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(pids.List)).To(BeNumerically(">", 2))
	})

	It("proc io", func() {
		io := ProcIo{}
		err := io.Get(os.Getppid())
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("proc state", func() {
		state := ProcState{}
		err := state.Get(os.Getppid())
		if runtime.GOOS != "darwin" {
			Expect(err).ToNot(HaveOccurred())
			if runtime.GOOS == "linux" {
				Expect([]RunState{RunStateRun, RunStateSleep}).To(ContainElement(state.State))
			}
			Expect([]string{"go", "go.exe", "ginkgo"}).To(ContainElement(state.Name))
		} else {
			Expect(err).To(Equal(ErrNotImplemented))
		}

		err = state.Get(invalidPid)
		if runtime.GOOS != "darwin" {
			Expect(err).To(HaveOccurred())
		} else {
			Expect(err).To(Equal(ErrNotImplemented))
		}
	})

	It("proc mem", func() {
		mem := ProcMem{}
		err := mem.Get(os.Getppid())
		Expect(err).ToNot(HaveOccurred())

		err = mem.Get(invalidPid)
		Expect(err).To(HaveOccurred())
	})

	It("proc time", func() {
		time := ProcTime{}
		err := time.Get(os.Getppid())
		Expect(err).ToNot(HaveOccurred())
		Expect(time.User).To(BeNumerically(">", 0))
		Expect(time.Sys).To(BeNumerically(">", 0))

		err = time.Get(invalidPid)
		Expect(err).To(HaveOccurred())
	})

	It("proc args", func() {
		args := ProcArgs{}
		err := args.Get(os.Getppid())
		Expect(err).ToNot(HaveOccurred())
		Expect(len(args.List)).To(BeNumerically(">=", 1))
	})

	It("proc exe", func() {
		exe := ProcExe{}
		err := exe.Get(os.Getppid())
		Expect(err).ToNot(HaveOccurred())
		Expect([]string{"go", "go.exe", "ginkgo"}).To(ContainElement(filepath.Base(exe.Name)))
	})

	It("disk list", func() {
		disk := DiskList{}
		err := disk.Get()
		if runtime.GOOS == "darwin" {
			Expect(err).To(Equal(ErrNotImplemented))
		} else {
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("system info", func() {
		info := SystemInfo{}
		err := info.Get()
		Expect(err).ToNot(HaveOccurred())
	})

	It("system distribution", func() {
		dist := SystemDistribution{}
		err := dist.Get()
		Expect(err).ToNot(HaveOccurred())
	})

	It("returns NetConnState string", func() {
		Expect(ConnStateEstablished.String()).To(Equal("established"))
		Expect(ConnStateSynSent.String()).To(Equal("syn_sent"))
		Expect(ConnStateSynRecv.String()).To(Equal("syn_recv"))
		Expect(ConnStateFinWait1.String()).To(Equal("fin_wait1"))
		Expect(ConnStateFinWait2.String()).To(Equal("fin_wait2"))
		Expect(ConnStateTimeWait.String()).To(Equal("time_wait"))
		Expect(ConnStateClose.String()).To(Equal("close"))
		Expect(ConnStateCloseWait.String()).To(Equal("close_wait"))
		Expect(ConnStateLastAck.String()).To(Equal("last_ack"))
		Expect(ConnStateListen.String()).To(Equal("listen"))
		Expect(ConnStateClosing.String()).To(Equal("closing"))
	})

	It("returns NetConnProto string", func() {
		Expect(ConnProtoUdp.String()).To(Equal("udp"))
		Expect(ConnProtoTcp.String()).To(Equal("tcp"))
		Expect(ConnProtoRaw.String()).To(Equal("raw"))
	})
})
