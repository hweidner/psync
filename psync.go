// Copyright 2018-2020 by Harald Weidner <hweidner@gmx.net>. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// Version 3 that can be found in the LICENSE.txt file.

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"syscall"
	"time"
)

// BUFSIZE defines the size of the buffer used for copying. It is currently 64kB.
const BUFSIZE = 64 * 1024

// Buffer, Channels and Synchronization
var (
	buffer [][BUFSIZE]byte
	dch    = make(chan string, 100) // dispatcher channel - get work into work queue
	wch    = make(chan string, 100) // worker channel - get work from work queue to copy thread
	wg     sync.WaitGroup           // waitgroup for work queue length
)

// Commandline options and parameters
var (
	src, dest            string // source and destination directory
	optThreads           uint   // number of threads
	optVerbose, optQuiet bool   // verbose and quiet flags
	optTimes, optOwner   bool   // preserve timestamps and owner flag
	optCreate            bool   // create destination directory flag
	optSync              bool   // sync mode
)

func main() {
	// parse commandline flags
	flags()

	// check or create the destination directory
	prepareDestDir()

	// clear umask, so that it does not interfere with explicite permissions
	// used in os.FileOpen()
	syscall.Umask(0000)

	// initialize buffers
	buffer = make([][BUFSIZE]byte, optThreads)

	// Start dispatcher and copy threads
	go dispatcher()
	for i := uint(0); i < optThreads; i++ {
		go copyDir(i)
	}

	// start copying top level directory
	wg.Add(1)
	dch <- ""

	// wait for work queue to get empty
	wg.Wait()
}

// Function flags parses the command line flags and checks them for sanity.
func flags() {
	flag.UintVar(&optThreads, "threads", 16, "Number of threads to run in parallel")
	flag.BoolVar(&optVerbose, "verbose", false, "Verbose mode")
	flag.BoolVar(&optQuiet, "quiet", false, "Quiet mode")
	flag.BoolVar(&optTimes, "times", false, "Preserve time stamps")
	flag.BoolVar(&optOwner, "owner", false, "Preserve user/group ownership (root only)")
	flag.BoolVar(&optCreate, "create", false, "Create destination directory, if needed (with standard permissions)")
	flag.BoolVar(&optSync, "sync", false, "Run in sync mode, copy only files that do not exist on destination side (VERY LIMITED, USE WITH CARE)")
	flag.Parse()

	if flag.NArg() != 2 || flag.Arg(0) == "" || flag.Arg(1) == "" || optThreads > 1024 {
		usage()
	}

	if optThreads == 0 {
		optThreads = 16
	}
	src = flag.Arg(0)
	dest = flag.Arg(1)
}

// Function usage prints a message about how to use psync, and exits.
func usage() {
	fmt.Println("Usage: psync [options] source destination")
	flag.Usage()
	os.Exit(1)
}

// Function prepareDestDir checks for the existence of the destination,
// or creates it if the flag '-create' is set.
func prepareDestDir() {
	if optCreate {
		// create destination directory
		err := os.MkdirAll(dest, os.FileMode(0777))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR - unable to create destination dir %s: %s\n", dest, err)
			os.Exit(1)
		}
	} else {
		// test the existence of destination directory prior to syncing
		stat, err := os.Stat(dest)
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR - destination directory %s does not exist: %s.\nUse '-create' to create it.\n", dest, err)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR - cannot stat() destination directory %s: %s.\n", dest, err)
			os.Exit(1)
		}
		if !stat.IsDir() {
			fmt.Fprintf(os.Stderr, "ERROR - destination %s exists, but is not a directory\n", dest)
			os.Exit(1)
		}
	}
}

// Function dispatcher maintains a work list of potentially arbitrary size.
// Incoming directories (over the dispather channel) will be forwarded to a
// copy thread through the worker channel, or stored in the work list if no
// copy thread is available. For easier memory handling, the work list is
// treated last-in-first-out.
func dispatcher() {
	worklist := make([]string, 0, 1000)
	var dir string
	for {
		if len(worklist) == 0 {
			dir = <-dch
			worklist = append(worklist, dir)
		} else {
			select {
			case dir = <-dch:
				worklist = append(worklist, dir)
			case wch <- worklist[len(worklist)-1]:
				worklist = worklist[:len(worklist)-1]
			}
		}
	}
}

// Function copyDir receives a directory on the worker channel and copies its
// content from src to dest. Files are copied sequentially. If a subdirectory
// is discovered, it is created on the destination side, and then inserted into
// the work queue through the dispatcher channel.
func copyDir(id uint) {
	for {
		// read next directory to handle
		dir := <-wch
		if optVerbose {
			fmt.Printf("[%d] Handling directory %s%s\n", id, src, dir)
		}

		// read content of source directory
		files, err := ioutil.ReadDir(src + dir)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - could not read directory %s: %s\n", src+dir, err)
			}
			wg.Done()
			continue
		}

		// read content of destination directory, if needed
		desthash := make(map[string]os.FileInfo)
		if optSync {
			destfiles, err := ioutil.ReadDir(dest + dir)
			if err != nil {
				if !optQuiet {
					fmt.Fprintf(os.Stderr, "WARNING - could not read directory %s: %s\n", dest+dir, err)
				}
				// Skip whole directory as it is unsafe to continue in non-sync mode
				wg.Done()
				continue
			}
			// hash destination files for fast access to filenames
			for _, file := range destfiles {
				desthash[file.Name()] = file
			}
		}

		for _, f := range files {
			fname := f.Name()
			if fname == "." || fname == ".." {
				continue
			}

			if f.IsDir() {
				// entry is a directory. Create it on destination side, if needed
				if !optSync || desthash[fname] == nil || !desthash[fname].IsDir() {
					perm := f.Mode().Perm()
					err := os.Mkdir(dest+dir+"/"+fname, perm)
					if err != nil {
						if !optQuiet {
							fmt.Fprintf(os.Stderr, "WARNING - could not create directory %s: %s\n",
								dest+dir+"/"+fname, err)
						}
						continue
					}
				}

				// submit directory to work queue
				wg.Add(1)
				dch <- dir + "/" + fname
				continue
			}

			// Entry is a file, symbolic link, or special file. Copy file sequentially
			if !optSync || desthash[fname] == nil { // TODO: also copy if destination file exists but differs
				if optVerbose {
					fmt.Printf("[%d] Copying %s%s/%s to %s%s/%s\n",
						id, src, dir, fname, dest, dir, fname)
				}
				copyFile(id, dir+"/"+fname, f)
			}
		}

		// preserve ownership and/or timestamp destination directory
		finfo, err := os.Stat(src + dir)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - could not read fileinfo of directory %s: %s\n",
					dest+dir, err)
			}
		} else {
			// preserve user and group of the destination directory
			if optOwner {
				preserveOwner(dest+dir, finfo, "directory")
			}
			// setting the timestamps of the destination directory
			if optTimes {
				preserveTimes(dest+dir, finfo, "directory")
			}
		}
		if optVerbose {
			fmt.Printf("[%d] Finished directory %s%s\n", id, src, dir)
		}
		wg.Done()
	}
}

// Function copyFile copies a file from the source to the destination directory.
func copyFile(id uint, file string, f os.FileInfo) {
	mode := f.Mode()
	switch {

	case mode&os.ModeSymlink != 0: // symbolic link
		// read link
		link, err := os.Readlink(src + file)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - link %s disappeared while copying %s\n", src+file, err)
			}
			return
		}

		// write link to destination
		err = os.Symlink(link, dest+file)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - link %s could not be created: %s\n", dest+file, err)
			}
			return
		}

		// preserve owner of symbolic link
		if optOwner {
			preserveOwner(dest+file, f, "link")
		}
		// preserving the timestamps of links seems not be supported in Go
		// TODO: it should be possible by using the futimesat system call,
		// see https://github.com/golang/go/issues/3951
		//if times {
		//	preserveTimes(dest+file, f, "link")
		//}

	case mode&(os.ModeDevice|os.ModeNamedPipe|os.ModeSocket) != 0: // special files
	// TODO: not yet implemented

	default:
		// copy regular file
		// open source file for reading
		rd, err := os.Open(src + file)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s disappeared while copying: %s\n", src+file, err)
			}
			return
		}
		defer rd.Close()

		// open destination file for writing
		perm := mode.Perm()
		wr, err := os.OpenFile(dest+file, os.O_WRONLY|os.O_CREATE, perm)
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s could not be created: %s\n", dest+file, err)
			}
			return
		}
		defer wr.Close()

		// copy data
		_, err = io.CopyBuffer(wr, rd, buffer[id][:])
		if err != nil {
			if !optQuiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s could not be created: %s\n", dest+file, err)
			}
			return
		}

		if optOwner {
			preserveOwner(dest+file, f, "file")
		}
		if optTimes {
			preserveTimes(dest+file, f, "file")
		}

	}
}

// Function preserveOwner transfers the ownership information from the source to
// the destination file/directory.
func preserveOwner(name string, f os.FileInfo, ftype string) {
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		uid := int(stat.Uid)
		gid := int(stat.Gid)

		var err error
		if ftype == "link" {
			err = syscall.Lchown(name, uid, gid)
		} else {
			err = os.Chown(name, uid, gid)
		}

		if err != nil && !optQuiet {
			fmt.Fprintf(os.Stderr, "WARNING - could not change ownership of %s %s: %s\n",
				ftype, name, err)
		}
	}
}

// Function preserveTimes transfers the access and modification timestamp from
// the source to the destination file/directory.
func preserveTimes(name string, f os.FileInfo, ftype string) {
	mtime := f.ModTime()
	atime := mtime
	if stat, ok := f.Sys().(*syscall.Stat_t); ok {
		atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	}
	err := os.Chtimes(name, atime, mtime)
	if err != nil && !optQuiet {
		fmt.Fprintf(os.Stderr, "WARNING - could not change timestamps for %s %s: %s\n",
			ftype, name, err)
	}
}
