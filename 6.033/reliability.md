Book sections 8.1, 8.2, and 8.3
----------
- RAID 1: Mirror data across 2 disks.
 - Pro: Can handle single-disk failures
 - Pro: Performance improvement on reads (issue two in parallel),
 not a terrible performance hit on writes (have to issue two
 writes, but you can issue them in parallel too)
 - Con: To mirror N disks' worth of data, you need 2N disks
 - RAID 4: With N disks, add an additional parity disk. Secto

Sector i on
 the parity disk is the XOR of all of the sector i's from the data
 disk.
 - Pro: Can handle single-disk failures (if one disk fails, xor
 the other disks to recover its data)
 - Can use same technique to recover from single-sector errors
 - Pro: To store N disks' worth of data we only need N+1 disks
 - Pro: Improved performance if you stripe files across the
 array. E.g., an N-sector-length file can be stored as one
sector
 per disk. Reading the whole file means N parallel 1-sector
reads
 instead of 1 long N-sector read.
 - RAID is a system for reliability, but we never forget about
 performance, and in fact performance influenced much of the
 design of RAID.
 - Con: Every write hits the parity disk.

Transaction
--------
Book sections 9.1, 9.2.1, and 9.2.2

- Solution: recover the disk after a crash
 - Clean up refcounts
 - Delete tmp files
 - (see slides for code)
 - What if we crash during recover? Then just run recover again
 after that crash.

Shadow copies work because they perform updates/changes on a copy
 and automatically install a new copy using an atomic operation
 (in this case, single-sector writes)
