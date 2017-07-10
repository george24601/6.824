Part A: Key/value service without log compaction
--------

1. implement as a replicated state machine. Use a channel to notify that state has been processed, so that the blocking request can reply. 

Remember give go channel some capacity, otherwise, it will block until the message has been consumed 

2. same log index may appear multiple times

3. use a simple seq number to ensure no duplicate append command

Part B: Key/value service with log compaction
----------

Section 7 of the extended Raft paper for outline

compare maxraftstate to persister.RaftStateSize()
Whenever your key/value server detects that the Raft state size is approaching this threshold, it should save a snapshot, and tell the Raft library that it has snapshotted, so that Raft can discard old log entries.
