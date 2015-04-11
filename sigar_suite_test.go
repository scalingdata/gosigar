package sigar_test

import (
	. "github.com/scalingdata/ginkgo"
	. "github.com/scalingdata/gomega"

	"testing"
)

func TestGosigar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gosigar Suite")
}
