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


Notes for 3B
---------
