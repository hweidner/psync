psync Development goals
=======================

Here is how the psync usage might eventally look like.

None of these options are fixed yet. Plans for new functions might be dropped,
build in implicitely (without an option), or get another option name. More
functions and options might appear.

	psync [-verbose|-quiet] [-dryrun] [-threads <num>] [-owner] [-times] [-create]
	      [-sync] [-delete] [-update] source destination

	-verbose        - verbose mode, prints the current workload to STDOUT
	-quiet          - quiet mode, suppress warnings
    -dryrun         - dry run, do not actually copy files or directories
	-threads <num>  - number of concurrent threads, 1 <= <num> <= 1024, default 16
	-owner          - preserve ownership (user / group)
	-times          - preserve timestamps (atime / mtime)
	-create         - create destination directory, if needed (with standard permissions)
	-sync           - copy in "sync" mode, e.g. copy only files and directories that
	                  do not yet exist on the destination side, or exist but differ in
	                  size or time stamp (mtime).
	-delete         - in sync mode, delete files and directories that exist on the
	                  destination side, but not the source side.
	-update         - in sync mode, do not copy a file when there is a file with the
	                  same name and newer timestamp (mtime) on the destination side.
	                  This option has no effect on directories.
	source          - source directory
	destination     - destination directory
