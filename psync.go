// Copyright 2018 by Harald Weidner <hweidner@gmx.net>. All rights reserved.
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
)

// Channels and Synchronization
var (
	dch = make(chan string, 100) // dispatcher channel - get work into work queue
	wch = make(chan string, 100) // worker channel - get work from work queue to copy thread
	wg  sync.WaitGroup           // waitgroup for work queue length
)

// Commandline Flags
var (
	threads        uint   // number of threads
	src, dest      string // source and destination directory
	verbose, quiet bool   // verbose and quiet flags
)

func main() {
	// parse commandline flags
	flags()

	// Start dispatcher and copy threads
	go dispatcher()
	for i := uint(0); i < threads; i++ {
		go copyDir(i)
	}

	// start copying top level directory
	wg.Add(1)
	dch <- ""
	wg.Wait() // wait for work queue to get empty
	// time.Sleep(10 * time.Second)
}

// Function flags parses the command line flags and checks them for sanity.
func flags() {
	flag.UintVar(&threads, "threads", 16, "Number of threads to run in parallel")
	flag.BoolVar(&verbose, "verbose", false, "Verbose mode")
	flag.BoolVar(&quiet, "quiet", false, "Quiet mode")
	flag.Parse()

	if flag.NArg() != 2 || flag.Arg(0) == "" || flag.Arg(1) == "" || threads > 1024 {
		usage()
	}

	if threads == 0 {
		threads = 16
	}
	src = flag.Arg(0)
	dest = flag.Arg(1)
}

// Funktion usage prints a message about how to use psync, and exits.
func usage() {
	fmt.Println("Usage: psync [options] source destination")
	flag.Usage()
	os.Exit(1)
}

// Function dispatcher maintains a work list of potentially arbitrary size.
// Incoming directories (over the dispather channel) will be forwarded to a
// copy thread through the worker channel, or stored in the work list if no
// copy thread is available.
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
// the work queue through the dispather channel.
func copyDir(id uint) {
	for {
		dir := <-wch
		if verbose {
			fmt.Printf("[%d] Handling directory %s%s\n", id, src, dir)
		}
		files, err := ioutil.ReadDir(src + dir)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - could not read directory %s: %s\n", src+dir, err)
			}
			wg.Done()
			return
		}

		for _, f := range files {
			fname := f.Name()
			if fname == "." || fname == ".." {
				continue
			}
			if f.IsDir() {
				err := os.Mkdir(dest+dir+"/"+fname, 0755)
				if err != nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "WARNING - could not create directory %s: %s\n",
							dest+dir+"/"+fname, err)
					}
					continue
				}
				wg.Add(1)
				dch <- dir + "/" + fname
			} else {
				if verbose {
					fmt.Printf("[%d] Copying %s%s/%s to %s%s/%s\n",
						id, src, dir, fname, dest, dir, fname)
				}
				copyFile(dir+"/"+fname, f.Mode())
			}
		}
		if verbose {
			fmt.Printf("[%d] Finished directory %s%s\n", id, src, dir)
		}
		wg.Done()
	}
}

// Function copyFile copies a file from the source to the destination directory.
func copyFile(file string, mode os.FileMode) {
	switch {

	case mode&os.ModeSymlink != 0: // symbolic link
		link, err := os.Readlink(src + file)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - link %s disappeared while copying %s\n", src+file, err)
			}
			return
		}
		err = os.Symlink(link, dest+file)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - link %s could not be created: %s\n", dest+file, err)
			}
			return
		}

	case mode&(os.ModeDevice|os.ModeNamedPipe|os.ModeSocket) != 0: // special files
	// TODO: not yet implemented

	default:
		// copy regular file
		// open source file for reading
		rd, err := os.Open(src + file)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s disappeared while copying: %s\n", src+file, err)
			}
			return
		}
		defer rd.Close()

		// open destination file for writing
		wr, err := os.Create(dest + file)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s could not be created: %s\n", dest+file, err)
			}
			return
		}
		defer wr.Close()

		// copy data
		_, err = io.Copy(wr, rd)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "WARNING - file %s could not be created: %s\n", dest+file, err)
			}
			return
		}
	}
}
