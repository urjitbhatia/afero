// General Test structure:
// Create the source file system and write a bunch of files
// Create the destination file system
// Let filesync make them look identical
package sshfs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/classdojo/cyclotron/src/sshfs"
)

var _ = Describe("SSH FS tests", func() {
	It("Makes a directory", func() {
		fs, err := sshfs.NewSSHFS("10.1.3.244", 22, "/tmp/")
		Expect(err).To(BeNil())

		fs.Mkdir("is_this_crazy_talk", 0644)
	})
})
