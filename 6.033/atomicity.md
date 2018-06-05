
```
// find var’s last committed value
read(log, var):
    commits = {}
     // scan backwards
    for record r in log[len(log) - 1] .. log[0]:
     // keep track of commits
       if r.type == commit:
          commits.add(r.tid)
       if r.type == update and
          (r.tid in commits or r.tid == current_tid) and
          r.var == var:
           return r.new_value
read(var):
    return cell_read(var)

write(var, value):
    log.append(current_tid, update, var, read(var), value) cell_write(var, value)

recover(log):
  commits = {}
  for record r in log[len(log)-1] .. log[0]:
    if r.type == commit:
      commits.add(r.tid)
    if r.type == update and r.tid not in commits:
      cell_write(r.var, r.old_val) // undo

  for record r in log[0] .. log[len(log)-1]:
    if r.type == update and r.tid in commits:
      cell_write(r.var, r.new_value) // redo: we need this stage because we are writing to the volatile cache
```

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
	- What if we crash during recovery? No worries; recover() is
	idempotent. Can do it repeatedly.


To truncate

- Assuming no pending actions
         - Flush all cached updates to cell storage
         - Write a CHECKPOINT record
         - Truncate the log prior to the CHECKPOINT record
           - Usually amounts to deleting a file
       - With pending actions, delete before the checkpoint and
             earliest undecided record.

 - ABORT records
     - Can be used to help recovery and skip undo-ing aborted
       transaction.  Not necessary for correctness -- can always just
       pretend we crashed -- but can help.



Handson answers
------------

1. Because trx 2 is not commited

2. A with balance 900, and C with balance 3100

3. trx 3 is not marked as end. Therefore, the change is not flushed to disk

4. Exactly same, because no new write operations

5. Yes, because the log is properly redone

6. Winner: commited but not applied, to be redone. Loser: to be undone

7. Because 2 is at the pending state

8. 7 lines are rolled back. Trx 1 no longer needs redo

9. Yes. Idempotency



