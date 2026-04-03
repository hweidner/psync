// Copyright 2018-2020 by Harald Weidner <hweidner@gmx.net>. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// Version 3 that can be found in the LICENSE.txt file.

/*
Parallel Sync - parallel recursive copying of directories

psync is a tool which copies a directory recursively to another directory.
Unlike "cp -r", which walks through the files and subdirectories in sequential
order, psync copies several files concurrently by using threads.

See the [psync project on Github](https://github.com/hweidner/psync) for more
informations.

# Usage

psync is invoked as follows:

	psync [-verbose|-quiet] [-threads <num>] [-owner] [-times] [-create] source destination

	-verbose        - verbose mode, prints the current workload to STDOUT
	-quiet          - quiet mode, suppress warnings
	-threads <num>  - number of concurrent threads, 1 <= <num> <= 1024, default 16
	-owner          - preserve ownership (user / group)
	-times          - preserve timestamps (atime / mtime)
	-create         - create destination directory, if needed (with standard permissions)
	-sync           - sync mode, create directories or copy files that do not already exist

	source          - source directory
	destination     - destination directory

# Example

Copy all files and subdirectories from /data/src into /data/dest.

	psync -threads 8 /data/src /data/dest

/data/src and /data/dest must exist and must be directories.

# License

psync is released under the terms of the GNU General Public License, version 3.
*/
package main
