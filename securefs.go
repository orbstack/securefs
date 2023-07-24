package securefs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/orbstack/securefs/internal/syncx"
	"golang.org/x/sys/unix"
)

var (
	onceDefaultFS syncx.Once[*FS]
)

type FS struct {
	root string
	dfd  int
}

func NewFS(root string) (*FS, error) {
	dfd, err := unix.Open(root, unix.O_DIRECTORY|unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	return &FS{
		root: root,
		dfd:  dfd,
	}, nil
}

func Default() *FS {
	return onceDefaultFS.Do(func() *FS {
		fs, err := NewFS("/")
		if err != nil {
			panic(err)
		}

		return fs
	})
}

func (fs *FS) Close() error {
	return unix.Close(fs.dfd)
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// openat2 RESOLVE_IN_ROOT - so symlinks still work
	for {
		how := unix.OpenHow{
			Flags:   uint64(flag) | unix.O_CLOEXEC,
			Mode:    uint64(perm),
			Resolve: unix.RESOLVE_IN_ROOT,
		}
		fd, err := unix.Openat2(fs.dfd, name, &how)
		if err != nil {
			// need to check for EINTR - Go issues 11180, 39237
			// also EAGAIN in case of unsafe race
			if err == unix.EINTR || err == unix.EAGAIN {
				continue
			} else {
				return nil, err
			}
		}

		return os.NewFile(uintptr(fd), name), nil
	}
}

func (fs *FS) Open(name string) (*os.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *FS) Create(name string) (*os.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func (fs *FS) ReadFile(name string) ([]byte, error) {
	f, err := fs.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func (fs *FS) WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := fs.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (fs *FS) openParentOf(name string) (*os.File, error) {
	parentPath := filepath.Dir(name)
	return fs.OpenFile(parentPath, unix.O_DIRECTORY|unix.O_PATH, 0)
}

func (fs *FS) Remove(name string) error {
	// tricky: we have to open the *parent*, then unlinkat
	// unlinkat has no RESOLVE_IN_ROOT, AT_EMPTY_PATH, or AT_SYMLINK_NOFOLLOW
	parentDfile, err := fs.openParentOf(name)
	if err != nil {
		return err
	}
	defer parentDfile.Close()

	err = unix.Unlinkat(int(parentDfile.Fd()), filepath.Base(name), 0)
	if err != nil {
		// try rmdir like Go
		return unix.Unlinkat(int(parentDfile.Fd()), filepath.Base(name), unix.AT_REMOVEDIR)
	}

	return nil
}

func (fs *FS) Symlink(oldname, newname string) error {
	// same as above: open the new parent
	parentDfile, err := fs.openParentOf(newname)
	if err != nil {
		return err
	}
	defer parentDfile.Close()

	err = unix.Symlinkat(oldname, int(parentDfile.Fd()), filepath.Base(newname))
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) Mkdir(name string, perm os.FileMode) error {
	// same as above: open the new parent
	parentDfile, err := fs.openParentOf(name)
	if err != nil {
		return err
	}
	defer parentDfile.Close()

	err = unix.Mkdirat(int(parentDfile.Fd()), filepath.Base(name), uint32(perm))
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) MkdirAll(path string, perm os.FileMode) error {
	// end of recursion
	if path == "" || path == "." || path == "/" {
		return nil
	}

	// try first
	err := fs.Mkdir(path, perm)
	if err == nil || errors.Is(err, unix.EEXIST) {
		return nil
	}

	// if it failed, try w/ parent
	err = fs.MkdirAll(filepath.Dir(path), perm)
	if err != nil {
		return err
	}

	// try again
	err = fs.Mkdir(path, perm)
	if err != nil {
		if errors.Is(err, unix.EEXIST) {
			return nil
		}
		return err
	}

	return nil
}

func (fs *FS) ReadDir(name string) ([]os.DirEntry, error) {
	f, err := fs.OpenFile(name, unix.O_DIRECTORY|unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.ReadDir(0)
}

func (fs *FS) Stat(name string) (os.FileInfo, error) {
	f, err := fs.OpenFile(name, unix.O_PATH, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Stat()
}

func (fs *FS) ResolvePath(name string) (string, error) {
	file, err := fs.OpenFile(name, unix.O_PATH, 0)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// from magic link
	return os.Readlink(fmt.Sprintf("/proc/self/fd/%d", file.Fd()))
}

// quick functions
func OpenFile(at string, name string, flag int, perm os.FileMode) (*os.File, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.OpenFile(name, flag, perm)
}

func Open(at string, name string) (*os.File, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.Open(name)
}

func Create(at string, name string) (*os.File, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.Create(name)
}

func ReadFile(at string, name string) ([]byte, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.ReadFile(name)
}

func WriteFile(at string, name string, data []byte, perm os.FileMode) error {
	fs, err := NewFS(at)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.WriteFile(name, data, perm)
}

func Remove(at string, name string) error {
	fs, err := NewFS(at)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.Remove(name)
}

func Symlink(at string, oldname, newname string) error {
	fs, err := NewFS(at)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.Symlink(oldname, newname)
}

func Mkdir(at string, name string, perm os.FileMode) error {
	fs, err := NewFS(at)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.Mkdir(name, perm)
}

func MkdirAll(at string, path string, perm os.FileMode) error {
	fs, err := NewFS(at)
	if err != nil {
		return err
	}
	defer fs.Close()

	return fs.MkdirAll(path, perm)
}

func ReadDir(at string, name string) ([]os.DirEntry, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.ReadDir(name)
}

func ResolvePath(at string, name string) (string, error) {
	fs, err := NewFS(at)
	if err != nil {
		return "", err
	}
	defer fs.Close()

	return fs.ResolvePath(name)
}

func Stat(at string, name string) (os.FileInfo, error) {
	fs, err := NewFS(at)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	return fs.Stat(name)
}
