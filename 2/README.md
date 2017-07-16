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
1. The stale term - status check must apply to ALL request/responses
2. In 2a, we just need to worry about AppendEntries with empty slice as data payload 
3. The main instance should be stared in a goroutine, and run in an infinite loop
4. You need a way to mark the status of each raft instance: is it a follower, candidate, or leader?
5. Avoid RPC calls within lock. Release the mutex before them.
6. Every time you acquire the mutex, you need to check if the current instance state is consist, e.g., is the instance in a correct status?
7. To implmenet election timeout 

```
for {
	select {
		case <- heartbeatChannel: //your RPC call should feed into this channel
		case <- time.After(electionTimeout):{
				//Your code to change the instance status to candidate
			break
						    }
	}
}
```
8. To generate random time out
```
func randomElectionTimeout() time.Duration {
	return time.Duration(250 + rand.Intn(500) ) * time.Millisecond
}

```


Notes for 2B
--------
1. Applying the command should be on its own goroutine. Don't put it in the main run loop
2. If there is no new command to apply, use 
```
time.Sleep(10 * time.Millisecond)
```
To sleep a bit between
3. Use https://thesquareplanet.com/blog/students-guide-to-raft/ to see how to implement the quick backoff. Otherwise, the final few tests may not pass
4. Raft is ASYNC. This means even if you want to start replicate process inside start(), it has to be on a separte go routine. However, you most likely don't need to try replicate inside start()


Notes for 2C
---------
1. Every time you unlock, check if you need to to persist

2. The test suite has a lot of failover cases. Most of the test failure is caused by subtle bugs in 2A and 2B 
