Study Questions
---------
1. Suppose we have the scenario shown in the Raft paper's Figure 7: a cluster of seven servers, with the log contents shown. The first server crashes (the one at the top of the figure), and cannot be contacted. A leader election ensues. For each of the servers marked (a), (d), and (f), could that server be elected? If yes, which servers would vote for it? If no, what specific Raft mechanism(s) would prevent it from being elected?

2. There are five Raft servers. S1 is the leader in term 17. S1 and S2 become
isolated in a network partition. S3, S4, and S5 choose S3 as the leader for term 18. S3, S4,and S5 commit and execute commands in term 18. S1 crashes and restarts, and S1 and S2
run many elections, so that their currentTerms are 50. S3 crashes but does not re-start, and an 
instant later the network partition heals, so that S1, S2, S4, and S5 can communicate. Could 
S1 become the next leader? How could that happen, or what prevents it from happening?
Your answer should refer to specific rules in the Raft paper, and explain how the rules apply
in this specific scenario. 

3. Suppose the leader waited for all servers (rather than just a majority) to have
matchIndex[i] >= N before setting commitIndex to N (at the end of Rules for Servers).
What specific valuable property of Raft would this change break?

4. Suppose the leader immediately advanced commitIndex to its last log entry
without waiting for responses to AppendEntries RPCs, and regardless of the values in
matchIndex. Describe a specific situation in which this change would lead to incorrect execution of the service being replicated by Raft.



Notes for 2A
---------

Notes for 2B
--------

Notes for 2C
---------

