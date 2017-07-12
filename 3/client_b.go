package raftkv

import "labrpc"
import "crypto/rand"
import "math/big"
import	"sync"
import "time"
//import "log"

type Clerk struct {
	mu      sync.Mutex
	servers []*labrpc.ClientEnd
	// You will have to modify this struct.
	lastLeaderI int
	cID int64
	seqN int
	seenSeqNSet map[int]bool
	seenSeqN int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.lastLeaderI = 0
	ck.cID = nrand()
	ck.seqN = 0
	ck.seenSeqN = -1
	ck.seenSeqNSet = make(map[int]bool)

	//sleep a bit to ensure Raft instance finished one election
//	time.Sleep(150 * time.Millisecond)
	return ck
}

func (ck *Clerk) Get(key string) string {
	ck.mu.Lock()
	startI := ck.lastLeaderI
	nServers := len(ck.servers)
	curSeq := ck.seqN
	ck.seqN++
	ck.mu.Unlock()

	for {
		for i := 0; i < nServers; i++ {
			sI := (startI + i) % nServers;
			args := GetArgs{key, ck.cID, curSeq}
			reply := &GetReply{}
			ok := ck.servers[sI].Call("RaftKV.Get", &args, &reply)

			if ok {
				if reply.WrongLeader {
					//		log.Printf("Cient: %d is not leader", sI)
				}else {
					ck.mu.Lock()
					//log.Printf("Client GET: sets last known leader to %d", sI)
					ck.lastLeaderI = sI
					ck.mu.Unlock()

					return reply.Value
				}

			} else {
				//log.Printf("Cient call to %d failed", sI)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return ""
}

func (ck *Clerk) PutAppend(key string, value string, op string) {
	ck.mu.Lock()
	startI := ck.lastLeaderI
	nServers := len(ck.servers)
	curSeq := ck.seqN
	ck.seqN++
	seenSeqN := ck.seenSeqN
	ck.mu.Unlock()

	for {
		for i := 0; i < nServers; i++ {
			sI := (startI + i) % nServers;
			args := PutAppendArgs{key, value, op, ck.cID, curSeq, seenSeqN}
			reply := &GetReply{}

			//log.Printf("Client PA: %s -> %v to %d ", args.Key, args.Value, sI)
			ok := ck.servers[i].Call("RaftKV.PutAppend", &args, &reply)

			if ok {
				if reply.WrongLeader {
					//log.Printf("Client: %d is wrong leader", sI)
				}else {
					//					log.Printf("Client PA: sets last known leader to %d", sI)
					ck.mu.Lock()
					ck.lastLeaderI = sI

					ck.seenSeqNSet[curSeq] = true

					for ;; {

						_, ok = ck.seenSeqNSet[ck.seenSeqN + 1]

						if ok {
							ck.seenSeqN++
							delete(ck.seenSeqNSet, ck.seenSeqN)
						}else {
							break
						}
					}

					ck.mu.Unlock()
					return
				}

			} else {
				//log.Printf("Client call to %d is not OK", sI)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "Append")
}
