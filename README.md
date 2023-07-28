# securefs

Go library for secure file system operations scoped to an arbitrary root directory on Linux, without chroot, mount namespaces, or other privileged features.

This uses the Linux-specific [openat2](https://man7.org/linux/man-pages/man2/openat2.2.html) syscall with `RESOLVE_IN_ROOT` to prevent symlink escapes and race conditions. Other solutions like [securejoin](https://github.com/cyphar/filepath-securejoin) are subject to race conditions.

Unlike `O_NOFOLLOW`, this supports all file system operations and works with symlinks (as long as they don't escape the specified root directory).
