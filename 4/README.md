Notes for 4A
--------
1. The shardmaster should be implemented as a replicated state machine, i.e., very similar to the kv server in part 3

Notes for 4B
--------
1. The recommended approach is to have each replica group use Raft to log not just the sequence of Puts, Appends, and Gets but also the sequence of reconfigurations. You will need to ensure that at most one replica group is serving requests for each shard at any one time.
