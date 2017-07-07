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
	LogI  int
	Term int
}

type RaftKV struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	maxraftstate int // snapshot if log grows this big

	kToV map[string]string
	lastPAI int
	lastPATerm int
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
	}

	DPrintf("Server %d logged Get %s at index %d, term %d", kv.me, args.Key, index, term)
	for {
		select  {
		case applyMsg := <-kv.applyCh:

			opToApply := applyMsg.Command.(Op)
			if applyMsg.Index <= kv.lastPAI {

				if opToApply.O == "Get" {
					DPrintf("!!!SHOULD NOT APPLY A GET")
				}

				DPrintf("Server %d to apply log at %d", kv.me, applyMsg.Index)
				kv.PAWithinLock(opToApply)
			}else{
				sameLogEntry := opToApply.K == op.K && opToApply.O == op.O
				//TODO:consistency check here

				if sameLogEntry{
					reply.WrongLeader = false
					keyedV, _ := kv.kToV[args.Key]
					reply.Value = keyedV

				}else{

					DPrintf("Server %d find that commited log is not the immediate GET request ", kv.me, applyMsg.Index)
					reply.WrongLeader = true
				}
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
	defer kv.mu.Unlock() //TODO: eariler unlock?

	op := Op{args.Key, args.Value, args.Op, -1, -1}

	//DPrintf("Server %d PA %s -> %s", kv.me, args.Key, args.Value)
	index, term, isLeader :=  kv.rf.Start(op)
	DPrintf("Lead Server %d logged %s -> %s at index %d, term %d", kv.me, args.Key, args.Value, index, term)

	reply.WrongLeader = !isLeader

	if isLeader {
		kv.lastPAI = index
		kv.lastPATerm = term
	}
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

	//sleep a bit to ensure Raft instance finished one election
	time.Sleep(200 * time.Millisecond)
	return kv
}
