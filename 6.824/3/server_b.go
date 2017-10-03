package raftkv

import (
	"encoding/gob"
	"labrpc"
	"log"
	"raft"
	"sync"
	"time"
	"bytes"
)

const Debug = 0

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
	iToReply map[int] chan Op

	lastIncludedIndex int
}


func (kv *RaftKV) StartLog(op Op) (chan Op, bool, int) {
	index, _, isLeader :=  kv.rf.Start(op)

	if !isLeader {
		return nil, false, index
	}

	kv.mu.Lock()

	_, ok :=  kv.iToReply[index]

	if ok {
		DPrintf("!!!Dup index channel")
	}else{
		//		DPrintf("lead Server %d make channel for %s %s -> %s at index %d, seq %d", kv.me, op.O, op.K, op.V, index, op.SeqN)
		kv.iToReply[index] = make(chan Op, 10)
	}

	replyChan, _ := kv.iToReply[index]

	kv.mu.Unlock()

	return replyChan, true, index
}


func (kv *RaftKV) Get(args *GetArgs, reply *GetReply) {
	op := Op{args.Key, "", "Get", args.CID, args.SeqN}

	replyChan, isLeader, index := kv.StartLog(op)

	reply.WrongLeader = !isLeader

	if replyChan == nil {
		return
	}

	select {
	case opApplied := <-replyChan :

		DPrintf("Lead KV Server %d received %s %s -> %s at index %d, seq %d",  kv.me, op.O, op.K, op.V, index, op.SeqN)
		reply.WrongLeader = !(opApplied.K == op.K && opApplied.O == op.O)
		reply.Value = opApplied.V

	case <-time.After(1000 * time.Millisecond):
		DPrintf("!!!TIMEOUT:lead server %d %s %s -> %s at index %d, seq %d", kv.me, op.O, op.K, op.V , index, op.SeqN)
		reply.WrongLeader = true
	}
}

func (kv *RaftKV) HasDupWithinLock(op Op) bool {
	seenSeqN, ok := kv.cToSeenSeqN[op.CID]

	if ok && seenSeqN >= op.SeqN {


		DPrintf("Duplicate: Lead Server %d for request: %s %s -> %s : server has processed all seq <= %d, the request is at seq %d, current value is %s  ", kv.me, op.O, op.K, op.V, seenSeqN, op.SeqN, kv.kToV[op.K])
		return true
	}

	return false
}


func (kv *RaftKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	op := Op{args.Key, args.Value, args.Op, args.CID, args.SeqN}

	replyChan, isLeader, index := kv.StartLog(op)

	reply.WrongLeader = !isLeader

	if replyChan == nil {
		return
	}


	select {
	case opApplied := <-replyChan :

		DPrintf("Lead Server %d received %s %s -> %s at index %d, seqN %d", kv.me, args.Op, args.Key, args.Value, index, args.SeqN)
		sameAtIndex := opApplied.K == op.K && opApplied.V == op.V && opApplied.O == op.O

		if !sameAtIndex {
			DPrintf("!!!!Different op found on server %d: was %s %s -> %s, is %s %s -> %s ",  kv.me, op.O, op.K, op.V, opApplied.O, opApplied.K, opApplied.V)
		}

		reply.WrongLeader = !sameAtIndex

	case <-time.After(1000 * time.Millisecond):
		DPrintf("!!!TIMEOUT:lead server %d %s %s -> %s at index %d, seq %d", kv.me, op.O, op.K, op.V , index, op.SeqN)
		reply.WrongLeader = true
	}
}

func (kv *RaftKV) Kill() {
	kv.rf.Kill()
}

func (kv *RaftKV) ApplyWithinLock(applyMsg raft.ApplyMsg) {
	op := applyMsg.Command.(Op)

	//DPrintf("Server %d: Msg %s %s -> %s at index %d logged, about to apply in memory",kv.me, op.O, op.K, op.V ,applyMsg.Index)
	curV, ok := kv.kToV[op.K]

	if op.O == "Get" {
		if !ok{
			curV = ""
		}

		op.V = curV
	}else {
		if !kv.HasDupWithinLock(op){
			if ok && op.O == "Append" {
				kv.kToV[op.K] = curV + op.V
			} else {
				kv.kToV[op.K] = op.V
			}


			DPrintf("KV Server %d to apply %s %s -> %s : value after apply %s ", kv.me, op.O, op.K, op.V, kv.kToV[op.K]) 
		}

	}

	replyChan, ok := kv.iToReply[applyMsg.Index] 

	if ok {
	}else {
		//	DPrintf("CREATE CHANNEL: Server %d for request: %s %s -> %s at index %d, seq %d notifing", kv.me, op.O, op.K, op.V, applyMsg.Index, op.SeqN)

		replyChan = make(chan Op, 10)
		kv.iToReply[applyMsg.Index] = replyChan
	}

	replyChan <- op

	kv.cToSeenSeqN[op.CID] = op.SeqN

	//DPrintf("Server %d for request: %s %s -> %s at index %d, seq %d notifing", kv.me, op.O, op.K, op.V, applyMsg.Index, op.SeqN)
}


func (kv *RaftKV) CreateSSWithinLock(lastIncludedIndex int) []byte {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(kv.kToV)
	e.Encode(kv.cToSeenSeqN)
	e.Encode(lastIncludedIndex)
	data := w.Bytes()
	return data
}

func (kv *RaftKV) ReadSSWithinLock(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		kv.lastIncludedIndex = -1
		kv.kToV = make(map[string] string)
		kv.cToSeenSeqN = make(map [int64] int)
	}else{
		kToV := make(map[string] string)
		cToSeenSeqN := make(map [int64] int)
		var lastIncludedIndex int

		r := bytes.NewBuffer(data)
		d := gob.NewDecoder(r)

		d.Decode(&kToV)
		d.Decode(&cToSeenSeqN)
		d.Decode(&lastIncludedIndex)

		if (lastIncludedIndex <= kv.lastIncludedIndex) {
			DPrintf("!!!snapshot we get at server %d, index %d is older than the one we have at %d?!", kv.me, lastIncludedIndex, kv.lastIncludedIndex)
		}else{
			kv.kToV = kToV
			kv.cToSeenSeqN = cToSeenSeqN
			kv.lastIncludedIndex = lastIncludedIndex
			//TODO: delete obsolete channels?
		}

	}
}

func (kv *RaftKV) Apply() {
	for {
		select {
		case applyMsg := <- kv.applyCh :
			kv.mu.Lock()
			if applyMsg.UseSnapshot {
				kv.ReadSSWithinLock(applyMsg.Snapshot)
			} else {
				if (applyMsg.Index <= kv.lastIncludedIndex) {

					DPrintf("Server %d skipping index %d, because its last index at %d", kv.me,applyMsg.Index, kv.lastIncludedIndex)

				} else {
					//DPrintf("Server %d: Msg %s %s -> %s at index %d logged, about to apply in memory",kv.me, op.O, op.K, op.V ,applyMsg.Index)
					kv.ApplyWithinLock(applyMsg)

					if kv.maxraftstate >= 0 &&  kv.maxraftstate * 9 < kv.rf.RaftStateSize() * 10  {
						DPrintf("Server %d at index %d snapshotting", kv.me,applyMsg.Index)
						data := kv.CreateSSWithinLock(applyMsg.Index)
						kv.rf.UpdateSS(applyMsg.Index, data)
					}
				}
			}

			kv.mu.Unlock()
		}
	}
}

func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *RaftKV {
	gob.Register(Op{})

	kv := new(RaftKV)
	kv.me = me
	kv.maxraftstate = maxraftstate
	kv.ReadSSWithinLock(persister.ReadSnapshot())

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)
	kv.iToReply = make(map[int]chan Op)


	go func() {
		kv.Apply()
	}()

	time.Sleep(10 * time.Millisecond)
	return kv
}
