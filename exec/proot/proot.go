/* lang: pt-BR
Se esse pacote ficar bom provavelmente exportarei para seu proprio modulo externo

Estou fazendo merda, mas espero implementar varias chamadas para o syscall,
e terei que ver oque fazer com FDs abertos pelo syscalls, mas espero implementar logo.

Espero logo implementar o mecanismo do syscall para esse modulo, odeio ficar dependedo do programas externos.

qualquer coisa foi usar o cgo para incorporar o proot no Golang, mas quero algo totalmento feito no Golang
*/

// Implements Proot in Golang
package proot

// chroot, mount --bind, and binfmt_misc without privilege/setup for Linux/Android directly from golang
type PRoot struct {
	// The specified path typically contains a Linux distribution where
	// all new programs will be confined.  The default rootfs is /
	// when none is specified, this makes sense when the bind mechanism
	// is used to relocate host files and directories.
	Rootfs Binding

	// This option makes any file or directory of the host rootfs
	// accessible in the confined environment just as if it were part of
	// the guest rootfs.
	//
	// "mount path" => Virtual/Host filesystem
	Binds map[string]Binding

	// Execute guest programs through QEMU as specified by command.
	//
	// Each time a guest program is going to be executed, PRoot inserts
	// the QEMU user-mode command in front of the initial request.
	// That way, guest programs actually run on a virtual guest CPU
	// emulated by QEMU user-mode.  The native execution of host programs
	// is still effective and the whole host rootfs is bound to
	// /host-rootfs in the guest environment.
	Qemu string

	// Make current kernel appear as kernel release.
	//
	// If a program is run on a kernel older than the one expected by its
	// GNU C library, the following error is reported: "FATAL: kernel too
	// old".  To be able to run such programs, PRoot can emulate some of
	// the features that are available in the kernel release specified by
	// *string* but that are missing in the current kernel.
	KernelRelease string

	// Make current user and group.
	//
	// This option makes the current user and group appear as uid and
	// gid.  Likewise, files actually owned by the current user and
	// group appear as if they were owned by uid and gid instead.
	UID, GID int

	// Map ports to others.
	//
	// This option makes PRoot intercept bind and connect system calls,
	// and change the port they use. The port map is specified
	// with the syntax: -b *port_in*:*port_out*. For example,
	// an application that runs a MySQL server binding to 5432 wants
	// to cohabit with other similar application, but doesn't have an
	// option to change its port. PRoot can be used here to modify
	// this port: proot -p 5432:5433 myapplication. With this command,
	// the MySQL server will be bound to the port 5433.
	// This command can be repeated multiple times to map multiple ports.
	Port map[int16]int16

	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	// If Env is nil, the new process uses the current process's
	// environment.
	// If Env contains duplicate environment keys, only the last
	// value in the slice for each duplicate key is used.
	Env []string

	// Set the initial working directory.
	//
	// Some programs expect to be launched from a given directory but do
	// not perform any chdir by themselves.  This option avoids the
	// need for running a shell and then entering the directory manually.
	Dir string

	// Command and args to execute programer
	Command []string
}
