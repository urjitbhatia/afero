package sshfs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSSHFS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSH FS Suite")
}
