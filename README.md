Parallel Sync - parallel recursive copying of directories
=========================================================

psync is a tool which copies a directory recursively to another directory.
Unlike "cp -r", which walks through the files and subdirectories in sequential
order, psync copies several files concurrently by using threads.

Usage
-----

psync is invoked as follows:

	psync [-verbose] [-threads <num>] source destination
	
	-verbose        - verbose mode, prints the current workload to STDOUT
	-quiet          - quiet mode, suppress warnings
	-threads <num>  - number of concurrent threads, 1 <= <num> <= 1024, default 16
	source          - source directory
	destination     - destination directory

Example
-------

Copy all files and subdirectories from /data/src into /data/dest.

	psync -threads 8 /data/src /data/dest

/data/src and /data/dest must exist and must be directories.

Why should I use it
-------------------

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
------------------------------

Parallel copying is typically not so useful when copying between local or
very fast hard disks. psync can be used there, and with a moderate concurrency
level like 2..5, it can be slightly faster than a sequential copy.

Parallel copying should never be used when duplicating directories on the same
physical hard disk. Even sequential copies suffer from the frequent hard disk head
movements which are needed to read and write concurrently on/to the same disk.
Parallel copying even increases the amount of head movements.

How it works
------------

psync uses goroutines for copying files in parallel. By default, 16 copy workers
are spawned as goroutines, but the number can be adjusted with the -threads switch.

Each worker waits for a directory to be submitted. It then handles all the
directory entries sequentially. Files are copied one by one to the destination
directory. When subdirectories are discovered, they are created on the destination
side. Traversal of the subdirecory is then submitted to other workers and thus done
in parallel to the current workload.

Limits and TODOs
----------------

psync currently can only handle directories, regular files, and symbolic links.
Other filesystem entries like devices, sockets or named pipes are silently ignored.

psync preserves the Unix permissions (rwx) of the copied files and directories,
but has currently no support for preserving other permission bits (suid, sticky),
ownership, or timestamps. Destination files are created "as is", according to the
current user and group.

psync does currently implement a simple recursive copy, like "cp -r", and not
a versatile sync algorithm like rsync. There is no check wether a file already
exists in the destination, nor its content and timestamps. Existing files on the
destination side are not deleted when they don't exist on the source side.

License
-------

psync is released under the terms of the GNU General Public License, version 3.
