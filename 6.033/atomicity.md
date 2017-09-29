How to use a log for transactions
- On begin: allocate new transaction ID (TID)
	- On write: append entry to log
	- On read: scan log to find last committed value
	- On commit: write commit record
	- This is the commit point
	- Atomic because we can assume it's a single-sector write
	- Another way to do it would be to put checksums on each record
	and ignore partially-written records

	- How to recover
	- Scan the log backwards, determine what actions aborted, and
	undo them
- (see slide for code)
	- What if we crash during recovery? No worries; recover() is
	idempotent. Can do it repeatedly.

