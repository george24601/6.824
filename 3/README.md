Study Questions
----------
1. Ben Bitdiddle notices that his Get operations are too slow. Given that they
don’t actually modify state, he figures he’ll be safe if leaders service them immediately without
putting them through the log. Describe a scenario in which this change causes incorrect
behavior.

2. Could a received InstallSnapshot RPC cause the state machine to go backwards in time? That is, could step 8 in Figure 13 cause the state machine to be reset so that it reflects fewer executed operations? If yes, explain how this could happen. If no, explain why it can't happen.

3. Ben is having some trouble passing tests and he asks you for help. You
notice that, to avoid submitting too many operations to the Raft log, Ben is checking for
duplicate requests in his RPC handlers but not when applying operations from the apply
channel. Could this cause an operation to be committed twice? If not, explain. If so, please
give a sequence of events exhibiting this bad behavior.



Notes for 3A
--------
1. See https://thesquareplanet.com/blog/students-guide-to-raft/ on how to solve 3A
2. For client, sleep a bit between each iteration of trying all kv servers, to handle the case of election in progress
3. Though raft is asynch, our kv server is synch in the eyes of client. This means for each request, we should create a channel listening to the reply returned from the replicated state machine. The reply should also has all the kv-data we need to pass back to client. Cleaning up these callback channels is nice but not necessary
4. Add a long timeout,e.g., 1 sec,  when listening to callback channels, and mark the current request is bad. Otherwise, the handler will hang there
5. 3A should not need much if any change to the raft code

Notes for 3B
---------
1. Any time raft server needs to install a snapshot, you most likely need to persist the snapshot state. Otherwise, recovery will be using a stale snapshot 
2. Any time raft server refers to an index that is potentialy no greater than the last index included in the snapshot, you need to either return early or find that is an error case
3. In kv server and raft, when we get command to use a snapshot, we need to check to make sure the snapshot is more stale from the current snapshot/state
4. The sides involved in InstallSnapshot RPC calls need to implement general Raft rules regarding inconsitent term too. 
