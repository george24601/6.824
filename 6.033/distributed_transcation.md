message failures solved with reliable transport protocol (sequence numbers + ACKs)

if workers fail after the commit point, we cannot abort the transaction. workers MUST BE able to recover into a prepared state

Workers write PREPARE records once prepared. the recovery process — reading through the log — will indicate which transactions are prepared but not committed

Coordinator failure during prepare: coordinator recovers and aborts on all sides. However, if the coordniator crashed in the commit phase, it should keep committing

Note that 2PC guarantees atomicity but not consistency!

"Prepared" means that the workers will definitely commit even if they crash.

 Worker failure before prepare: coordinator sends abort messages to all workers and the client, and writes and ABORT record on its log. Upon recovery, the worker will find that this transaction has aborted 

After the commit point,  Lost commit message: coordinator times out and resends. Worker
 will also send a request for the status of this particular transaction.

 Worker failure before receiving commit: can't abort now!
 - After receive prepare messages, workers write PREPARE records into their log.
 - On recovery, they scan the log to determine what transactions are prepared but not yet committed or aborted.
 - They then make a request to the server asking for the status of that transaction. In this case, it has committed, so the server will send back a commit message.

Coordinator typically keeps a table mapping transaction ID to its state, for quick lookup here.

Once coordinator has heard that all workers are prepared, it writes COMMIT to its own log. This is the commit point.
 - Once coordinator has heard that all workers are committed, it writes a DONE record to its own log. At that point, transaction
 is totally done.

Coordinator failure before prepare: on recovery, abort (send abort message to workers + client)
 Why not try to continue on with the transaction? Likely the client has also timed out and assumed abort. Aborting everywhere is much cleaner.

2PC can be impractical. Sometimes we use compensating actions instead (e.g., banks let you cancel a transaction for free if you
 do so within X minutes of initiating the transaction).

2PC does NOT guarantee that they agree at the same instant, nor that they even agree within bounded time. This is an instance Two-Generals Paradox


