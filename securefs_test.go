package securefs

import (
	"errors"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestOpenFile(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// open file
	f, err := OpenFile(root, "a/file", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	// write to file
	_, err = f.Write([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	// close file
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// read file
	d, err := os.ReadFile(root + "/a/file")
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != "world" {
		t.Fatal("expected world")
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// open file
	f, err := Open(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}

	// read file
	var buf [10]byte
	n, err := f.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatal("expected 5")
	}
	if string(buf[:n]) != "hello" {
		t.Fatal("expected hello")
	}

	// close file
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	f, err := Create(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	// check exists
	_, err = os.Stat(root + "/a/file")
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// read file
	d, err := ReadFile(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != "hello" {
		t.Fatal("expected hello")
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// write file
	err = WriteFile(root, "a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// read file
	d, err := os.ReadFile(root + "/a/file")
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != "hello" {
		t.Fatal("expected hello")
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// remove file
	err = Remove(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}

	// check exists
	_, err = os.Stat(root + "/a/file")
	if err == nil {
		t.Fatal("expected not exists")
	}

	// try dir
	err = Remove(root, "a")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSymlink(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// symlink file
	err = Symlink(root, "a/file", "a/file2")
	if err != nil {
		t.Fatal(err)
	}

	// read symlink
	d, err := os.Readlink(root + "/a/file2")
	if err != nil {
		t.Fatal(err)
	}

	if d != "a/file" {
		t.Fatal("expected a/file")
	}
}

func TestMkdir(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create dir
	err = Mkdir(root, "a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// check exists
	_, err = os.Stat(root + "/a")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMkdirAll(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create dir
	err = MkdirAll(root, "a/b", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// check exists
	_, err = os.Stat(root + "/a/b")
	if err != nil {
		t.Fatal(err)
	}

	// again
	err = MkdirAll(root, "a/b//.//", 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = MkdirAll(root, "/.//.a/b//.//", 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadDir(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create dir
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// read dir
	d, err := ReadDir(root, "a")
	if err != nil {
		t.Fatal(err)
	}

	if len(d) != 1 {
		t.Fatal("expected 1")
	}
	if d[0].Name() != "file" {
		t.Fatal("expected file")
	}
}

func TestStat(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create dir
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// stat file
	s, err := Stat(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}

	if s.Name() != "file" {
		t.Fatal("expected file")
	}
}

func TestResolvePath(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create dir
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// resolve path
	p, err := ResolvePath(root, "a/file")
	if err != nil {
		t.Fatal(err)
	}

	if p != root+"/a/file" {
		t.Fatal("expected root/a/file")
	}
}

func TestDefaultFS(t *testing.T) {
	t.Parallel()
	fs := Default()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	err = os.MkdirAll(root+"/a", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// create file
	err = os.WriteFile(root+"/a/file", []byte("hello"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// open file
	f, err := fs.Open(root + "/a/file")
	if err != nil {
		t.Fatal(err)
	}

	// read file
	var buf [10]byte
	n, err := f.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatal("expected 5")
	}
	if string(buf[:n]) != "hello" {
		t.Fatal("expected hello")
	}

	// close file
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAttemptSymlinkEscape(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	// create second root
	root2, err := os.MkdirTemp("", "securefs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root2)

	// create symlink
	err = os.Symlink(root2, root+"/a")
	if err != nil {
		t.Fatal(err)
	}

	// try write
	err = WriteFile(root, "a/file", []byte("hello"), 0600)
	if !errors.Is(err, unix.ENOENT) {
		t.Fatal("expected error")
	}

	// try read
	err = os.Symlink("/etc/passwd", root+"/afile")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ReadFile(root, "afile")
	if !errors.Is(err, unix.ENOENT) {
		t.Fatal("expected error")
	}

	// try stat
	_, err = Stat(root, "a")
	if !errors.Is(err, unix.ENOENT) {
		t.Fatal("expected error")
	}
}

func TestCloexec(t *testing.T) {
	t.Parallel()

	// open etc/passwd
	fs, err := NewFS("/etc")
	if err != nil {
		t.Fatal(err)
	}

	// open file
	f, err := fs.Open("passwd")
	if err != nil {
		t.Fatal(err)
	}

	// check cloexec
	flags, err := unix.FcntlInt(f.Fd(), unix.F_GETFD, 0)
	if err != nil {
		t.Fatal(err)
	}
	if flags&unix.FD_CLOEXEC == 0 {
		t.Fatal("expected cloexec")
	}

	// check dirfd cloexec
	flags, err = unix.FcntlInt(uintptr(fs.dfd), unix.F_GETFD, 0)
	if err != nil {
		t.Fatal(err)
	}
	if flags&unix.FD_CLOEXEC == 0 {
		t.Fatal("expected cloexec")
	}
}

func TestNonExistentRoot(t *testing.T) {
	t.Parallel()

	_, err := NewFS("/nonexistent")
	if !errors.Is(err, unix.ENOENT) {
		t.Fatal("expected error")
	}
}

func TestFileAsRoot(t *testing.T) {
	t.Parallel()

	_, err := NewFS("/etc/passwd")
	if !errors.Is(err, unix.ENOTDIR) {
		t.Fatal("expected error")
	}
}
