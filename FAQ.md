# Frequently Asked Questions

**Q: Why does psync handle whole directories in copy threads, not single files?**

**A:** The reason is to keep the communication overhead small.
Only directories need to be collected and distributed to the copy threads.

In theory, when copying flat directory structures, it can happen that a single
thread still has to copy many files, while the other threads already have finished.

In practise, psync is typically used on large directory structures with thousands
of files in hundrets of nested directories.
In the vast majority of time, there are more directories left than threads.
The workload gets inbalanced only at the very end of the process.

Moreover, even if psync would distribute each file to a copy thread, the same
inbalance could happen with a single large file. It would need to distribute chunks
of files, e.g. 64MB, but this increases the communication overhead even more.

**Q: Why is directory traversal done in each copy thread, not in one dedicated traversal thread?**

**A:** Because this puts a limit on concurrency.

If your directory/file sizes are in a range where copying is e.g. 5 times slower than traversing,
then a single traverse thread can satisfy at most 5 copy threads,
no matter what concurrency level is actually used.
This is especially an issue in sync mode, when most of the files do not have to be copied.

The design of psync prevents these limitations.
