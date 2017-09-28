By the end of section three, you should understand the differences between ordinary files, directories, and special files.
By the end of section four (along with section three), you should be able to explain what happens when a user opens a file. For instance, if a user opens /home/example.txt, what does the UNIX file system do in order to find the file's contents? You should understand this in detail (e.g., at the i-node level). As always, if you have any questions, post on Piazza!

How is its filesystem designed?

After reading section 5, you should understand the basics of processes in UNIX (e.g., how fork() works, how memory is shared, how processes communicate). After reading section 6, you should undersatnd the basics of the shell. For instance, you should be able to describe what happens if you type sh into the UNIX shell (how many processes would be running?).

What does the UNIX shell do?
How does it work?
Why is it useful?

---------
bounded buffers: virtualize communication links

 Attempt 3 (correct):
 - send(bb, message):
 acquire(bb.lock)
 while bb.in - bb.out = N:
 release(bb.lock) // repeatedly release and acquire, to allow
 acquire(bb.lock) // processes calling receive() to jump in
 bb.buf[bb.in mod N] <- message
 bb.in <- bb.in + 1
 release(bb.lock)
 return

- Fine-grained locking:
 - move(dir1, dir2, filename):
 acquire(dir1.lock)
 unlink(dir1, filename)
 release(dir1.lock)
 acquire(dir2.lock)
 link(dir2, filename)
 release(dir2.lock)
 - Better performance, but incorrect. What if dir2 is renamed
 between release and acquire?

 - Problem: need locks to implement locks
 - Solution: hardware support (atomic instructions)
 - x86 example: XCHG
 - XCHG reg, addr

i - acquire(lock):
 do:
 r <= 1
 XCHG r, lock
 while r == 1
 - Atomic operations made possible by the controller that manages
 access to memory

------
• Preemption
 Forces a thread to be interrupted so that we don’t have
 to rely on programmers correctly using yield().
 Requires a special interrupt and hardware support to
 disable other interrupts.
