package watcher

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/goromdb/testutil"
)

func TestVerifyFileMD5(t *testing.T) {
	type Case struct {
		file, md5file string
		expectedOK    bool
		expectError   bool
		subtest       string
	}
	cases := []Case{
		{"non-existing.txt", "non-existing.txt.md5", false, false, "non-existing file"},
		{"valid.txt", "non-existing.txt.md5", false, true, "non-existing md5 file"},
		{"valid.txt", "invalid-len.txt.md5", false, true, "invalid md5 length"},
		{"valid.txt", "invalid-sum.txt.md5", false, true, "invalid md5 sum"},
		{"valid.txt", "valid.txt.md5", true, false, "valid md5"},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			ok, err := verifyFileMD5(c.file, c.md5file)

			assert.Equal(t, ok, c.expectedOK)
			assert.Equal(t, err != nil, c.expectError)
		})
	}
}

func TestNewMD5Watcher(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)

	file := filepath.Join(dir, "hoge.txt")
	wcr := NewMD5Watcher(file, 1000, logger)

	assert.NotNil(t, wcr)
}

func TestStartMD5Watcher(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)

	file := filepath.Join(dir, "hoge.txt")
	wcr := NewMD5Watcher(file, 1000, logger)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	out := wcr.Start(ctx)
	cancel()
	<-out // out should get closed after cancel() call
}

func TestWatchMD5WatcherOutput(t *testing.T) {
	dir := testutil.CreateTmpDir()
	defer os.RemoveAll(dir)

	logbuf := new(bytes.Buffer)
	logger := log.New(logbuf, "", 0)

	file := filepath.Join(dir, "valid.txt")
	wcr := NewMD5Watcher(file, 100, logger)

	ctx, cancel := context.WithCancel(context.Background())
	out := wcr.Start(ctx)

	testutil.CopyFile(filepath.Join(dir, "valid.txt"), "valid.txt")
	testutil.CopyFile(filepath.Join(dir, "valid.txt.md5"), "valid.txt.md5")

	loadedFile := <-out
	cancel()
	<-out // out should get closed after cancel() call

	assert.Equal(t, file, loadedFile)

	_, err := os.Stat(filepath.Join(dir, "valid.txt.md5"))

	assert.False(t, os.IsExist(err))
}
