// Copyright 2018 by Harald Weidner <hweidner@gmx.net>. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// Version 3 that can be found in the LICENSE.txt file.

/*
Parallel Sync - parallel recursive copying of directories

psync is a tool which copies a directory recursively to another directory.
Unlike "cp -r", which walks through the files and subdirectories in sequential
order, psync copies several files concurrently by using threads.

Usage

psync is invoked as follows:

	psync [-verbose|-quiet] [-threads <num>] [-owner] [-times] source destination

	-create         - create a non existent destination dir with default permissions
	-interactive    - prompt before writing to non empty directory or for creating a missing destination dir"
	-no-clobber     - check destination to be empty before syncing (overrides -interactive option)
	-verbose        - verbose mode, prints the current workload to STDOUT
	-quiet          - quiet mode, suppress warnings
	-threads <num>  - number of concurrent threads, 1 <= <num> <= 1024, default 16
	-owner          - preserve ownership (user / group)
	-times          - preserve timestamps (atime / mtime)
	source          - source directory
	destination     - destination directory

Example

Copy all files and subdirectories from /data/src into /data/dest.

	psync -threads 8 /data/src /data/dest

/data/src and /data/dest must exist and must be directories.

Why should I use it

A recursive copy of a directory can be a throughput bound or latency bound
operation, depending on the size of files and characteristics of the source
and/or destination file system. When copying between standard file systems on
two local hard disks, the operation is typically throughput bound, and copying
in parallel has no performance advantage over copying sequentially. In this
case, you have a bunch of options, including "cp -r" or "rsync".

However, when copying from or to network file systems (NFS, CIFS), WAN storage
(WebDAV, external cloud services), distributed file systems (GlusterFS, CephFS)
or file systems that live on a DRBD device, the latency for each file access is
often limiting performance factor. With sequential copies, the operation can
consume lots of time, although the bandwidth is not saturated. In this case, it
can make up a significant performance boost if the files are copied in parallel.

Where psync should not be used

Parallel copying is typically not so useful when copying between local or
very fast hard disks. psync can be used there, and with a moderate concurrency
level like 2..5, it can be slightly faster than a sequential copy.

Parallel copying should never be used when duplicating directories on the same
physical hard disk. Even sequential copies suffer from the frequent hard disk head
movements which are needed to read and write concurrently on/to the same disk.
Parallel copying even increases the amount of head movements.

How it works

psync uses goroutines for copying files in parallel. By default, 16 copy workers
are spawned as goroutines, but the number can be adjusted with the -threads switch.

Each worker waits for a directory to be submitted. It then handles all the
directory entries sequentially. Files are copied one by one to the destination
directory. When subdirectories are discovered, they are created on the destination
side. Traversal of the subdirecory is then submitted to other workers and thus done
in parallel to the current workload.

Performance values

Here are some performance values comparing psync to cp and rsync when copying
a large directory structure with many small files from a local file system to
an NFS share.

The NFS server has an AMD E-350 CPU, 8 GB of RAM, a 2TB hard drive (WD Green
series) running Debian GNU/Linux 10 (Linux kernel 4.19). The NFS export is
a logical volume on the HDD with ext4 file system. The NFS export options are:
rw,no_root_squash,async,no_subtree_check.

The client is a workstation with AMD Ryzen 7 1700 CPU, 64 GB of RAM, running
Ubuntu 18.04 LTS with HWE stack (Linux kernel 5.3). The data to copy is located
on a 1TB SATA SSD with XFS, and buffered in memory. The NFS mount options are:
fstype=nfs,vers=3,soft,intr,async.

The hosts are connected over ethernet with 1 Gbit/s, ping latency is 160µs.

The data is an extracted linux kernel source code 4.15.2 tarball, containing
62273 files and 32 symbolic links in 4377 directories, summing up to 892 MB
(as seen by "du -s"). It is copied from the workstation to the server over NFS.

The options for the three commands are selected comparably. They copy the files
and links recursively and preserve permissions, but no ownership or time stamps.

    cp -r SRC DEST         1m50,288s   8,09 MB/s
    rsync -rl SRC/ DEST/   3m05,479s   4,81 MB/s
    psync SRC DEST         0m23,398s  38,12 MB/s

Limits and TODOs

psync currently can only handle directories, regular files, and symbolic links.
Other filesystem entries like devices, sockets or named pipes are silently ignored.

psync preserves the Unix permissions (rwx) of the copied files and directories,
but has currently no support for preserving other permission bits (suid, sticky).

When using the according options, psync tries to preserve the ownership
(user/group) and/or the access and modification time stamps. Preserve ownership
does only work when psync is running under the root user account. Preserving the
time stamps does only work for regular files and directories, not for symbolic
links.

psync does currently implement a simple recursive copy, like "cp -r", and not
a versatile sync algorithm like rsync. There is no check wether a file already
exists in the destination, nor its content and timestamps. Existing files on the
destination side are not deleted when they don't exist on the source side.

psync is being developed under Linux (Debian, Ubuntu, CentOS). It should work on
other distributions, but this has not been tested. It does currently not compile
for Windows, Darwin (MacOS), NetBSD and FreeBSD (but this should easily be
fixed.)

License

psync is released under the terms of the GNU General Public License, version 3.

*/
package main
