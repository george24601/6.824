package raftkv

import (
	"encoding/gob"
	"labrpc"
	"log"
	"raft"
	"sync"
	"time"
)

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}


type Op struct {
	K string
	V string
	O string
	CID  int64
	SeqN int
}

type RaftKV struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	maxraftstate int // snapshot if log grows this big

	kToV map[string]string
	cToSeenSeqN map[int64]int
	cToLoggedSeqNs map[int64]map[int] bool
}


func (kv *RaftKV) Get(args *GetArgs, reply *GetReply) {
	//enter every Get() in the Raft log with start
	kv.mu.Lock()
	defer kv.mu.Unlock() //TODO: eariler unlock?
	op := Op{args.Key, "", "Get", -1, -1}

	index, term, isLeader :=  kv.rf.Start(op)

	reply.WrongLeader = !isLeader

	if !isLeader {
		return
	} else {
		DPrintf("Server %d logged Get %s at index %d, term %d", kv.me, args.Key, index, term)
	}

	for {
		select  {
		case applyMsg := <-kv.applyCh:

			opToApply := applyMsg.Command.(Op)

			sameLogEntry := applyMsg.Index == index &&  opToApply.K == op.K && opToApply.O == op.O
			if sameLogEntry {
				keyedV, ok := kv.kToV[args.Key]

				if ok  {
					reply.Value = keyedV
				}else {
					reply.Value = ""
				}

				DPrintf("Lead Server %d returns %s -> %s at version of index %d", kv.me, args.Key, reply.Value, applyMsg.Index)

				return
			}else if applyMsg.Index < index {

				if opToApply.O == "Get" {
					//DPrintf("!!!Lead Server %d %s %s at index %d, term %d, current index at %d", kv.me, opToApply.O, opToApply.K, applyMsg.Index, term, index)
					//this may happen when we reply the log
				}else{
					kv.PAWithinLock(opToApply)
					DPrintf("Lead Server %d APPLIED %s %s -> %s at index %d, term %d", kv.me, opToApply.O, opToApply.K, opToApply.V, applyMsg.Index, term)
				}
			}else {

				DPrintf("!!!Should be wrong leader:Server %d find that commited log is not the immediate GET request ", kv.me, applyMsg.Index)

				return
			}
		}
	}
}


func (kv *RaftKV) PAWithinLock(op Op) {
	if op.O == "Get" {
		return
	}else {
		curV, ok := kv.kToV[op.K]

		if ok && op.O == "Append" {
			kv.kToV[op.K] = curV + op.V
		} else {
			kv.kToV[op.K] = op.V
		}
	}
}

func (kv *RaftKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	//enter every Get() in the Raft log with start
	kv.mu.Lock()

	seenSeqN, ok := kv.cToSeenSeqN[args.CID]

	if ok {
		if seenSeqN >= args.SeqN {

			DPrintf("Lead Server %d for request: %s %s -> %s : server has processed all seq <= %d, the request is at seq %d ", kv.me, args.Op, args.Key, args.Value, seenSeqN, args.SeqN)
			reply.WrongLeader = false
			kv.mu.Unlock()
			return 
		}
	}else {
		kv.cToSeenSeqN[args.CID] = -1
	}

	loggedSeqNs, exists := kv.cToLoggedSeqNs[args.CID]

	if exists {
		_, isDup:= loggedSeqNs[args.SeqN]

		if isDup {
			DPrintf("Lead Server %d for request: %s %s -> %s : server has processed the request is at seq %d ", kv.me, args.Op, args.Key, args.Value, args.SeqN)
			reply.WrongLeader = false
			kv.mu.Unlock()
			return 
		}
	}else {
		kv.cToLoggedSeqNs[args.CID] = make(map[int]bool)
	}


	op := Op{args.Key, args.Value, args.Op, args.CID, args.SeqN}
	//kv.mu.Unlock()

	//DPrintf("Server %d PA %s -> %s", kv.me, args.Key, args.Value)
	index, term, isLeader :=  kv.rf.Start(op)

	//kv.mu.Lock()
//	defer kv.mu.Unlock()

	if isLeader {
		DPrintf("Lead Server %d LOGGED %s %s -> %s at index %d, term %d", kv.me, args.Op, args.Key, args.Value, index, term)

		_, exists := kv.cToLoggedSeqNs[args.CID]

		if !exists {
			kv.cToLoggedSeqNs[args.CID] = make(map[int]bool)
		}

		kv.cToLoggedSeqNs[args.CID][args.SeqN] = true

		curSeenSeqN , _ :=  kv.cToSeenSeqN[args.CID] 

		if args.SeenSeqN > curSeenSeqN {
			kv.cToSeenSeqN[args.CID] = args.SeenSeqN

			for seqN := curSeenSeqN+1; seqN <= args.SeenSeqN; seqN++ {
				delete(kv.cToLoggedSeqNs[args.CID], seqN)
			}
		}
	}

	reply.WrongLeader = !isLeader

	kv.mu.Unlock()
}

//
// the tester calls Kill() when a RaftKV instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (kv *RaftKV) Kill() {
	kv.rf.Kill()
	// Your code here, if desired.
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots with persister.SaveSnapshot(),
// and Raft should save its state (including log) with persister.SaveRaftState().
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *RaftKV {
	// call gob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	gob.Register(Op{})

	kv := new(RaftKV)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	// You may need initialization code here.
	kv.kToV = make(map[string] string)
	kv.cToSeenSeqN = make(map [int64] int)
	kv.cToLoggedSeqNs = make(map [int64]map[int]bool)

	//sleep a bit to ensure Raft instance finished one election
	time.Sleep(50 * time.Millisecond)
	return kv
}
