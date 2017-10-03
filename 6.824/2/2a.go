package raft

import "sync"
import "labrpc"
import "time"
import "math/rand"

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	currentTerm int
	votedFor int
	status int
	fHB chan int 
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int
	var isleader bool
	term = rf.currentTerm
	isleader = rf.status == 2
	return term, isleader
}

func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

type RequestVoteArgs struct {
	Term int
	CandidateId int
}

type RequestVoteReply struct {
	Term int
	VoteGranted bool 
}

func (rf *Raft) CheckOldTermWithinLock(term int) {
	if(term > rf.currentTerm){
		rf.status = 0
		rf.currentTerm = term
		rf.votedFor = -1
	}
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		DPrintf("%d of status %d, term %d refused a vote from %d at term %d", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}

	rf.CheckOldTermWithinLock(args.Term)

	if rf.votedFor < 0 || rf.votedFor == args.CandidateId{ //TODO: log check

		DPrintf("%d of status %d, term %d grants a vote to %d at term %d", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)
		reply.VoteGranted = true
		reply.Term =  args.Term
		rf.votedFor = args.CandidateId

		if rf.status == 0{
			rf.fHB <- args.Term
		}
	}else{
		DPrintf("%d of status %d, term %d refuses to vote for %d at term %d", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)
		reply.VoteGranted = false
		reply.Term =  args.Term
	}
}

type AppendEntriesArgs struct {
	Term int
	LeaderId  int
}

type AppendEntriesReply struct {
	Term int
	Success bool 
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//TODO: better check with log terms later

	if args.Term < rf.currentTerm {
		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}

	rf.CheckOldTermWithinLock(args.Term)

	reply.Success = true
	reply.Term = args.Term

	DPrintf("%d of status %d, term %d agree to append entries for %d at term %d", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term)

	if(rf.status == 0){
		rf.fHB <- args.Term
	}else{
		DPrintf("!!!BAD STATE")
	}
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) BroadcastRV() {
}


func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

func (rf *Raft) Kill() {
	// Your code here, if desired.
}

func randomElectionTimeout() time.Duration {
	return time.Duration(rand.Intn(300) + 400) * time.Millisecond
}

func (rf *Raft) RunFollower() {
	DPrintf("Follower: %d" , rf.me)

	electionTimeout := randomElectionTimeout()

	for {
		select {
		case <- rf.fHB:

			//DPrintf("Follower(?) %d got a heartbeat" , rf.me)
			case <- time.After(electionTimeout):{
				rf.TryChangeStatus(0, 1)
				return
			}
		}
	}
}

func (rf *Raft) TryChangeStatus(from int, to int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.status == from {
		rf.status = to
	}
}

func (rf *Raft) RunCandidate() {
	var curTerm int
	rf.mu.Lock()
	rf.currentTerm= rf.currentTerm  + 1
	rf.votedFor = rf.me
	curTerm = rf.currentTerm
	rf.mu.Unlock()

	DPrintf("Candidate: %d at term: %d" , rf.me, curTerm)

	cRVR := make(chan RequestVoteReply, len(rf.peers))
	for  i:=0; i<len(rf.peers);i++ {
		if i == rf.me {
			continue
		}

		go func(k int) {
			args := &RequestVoteArgs{}
			args.CandidateId = rf.me
			args.Term = curTerm


			reply := &RequestVoteReply{ }
			ok := rf.sendRequestVote(k, args,  reply)

			if ok {
				cRVR <- *reply
			}else{
				DPrintf("Candidate %d to %d vote request is not ok, term: %d", rf.me, k, curTerm)
			}
		}(i)
	}


	var grantedVotes int = 1

	electionTimeout := randomElectionTimeout()


	var curStatus int
	rf.mu.Lock()
	curStatus = rf.status
	rf.mu.Unlock()

	if curStatus != 1 {
		DPrintf("Candidate state out of sync: %d", rf.me)
		return
	}

	for {
		select {
		case reply := <- cRVR:
			if reply.VoteGranted {

				grantedVotes++
				DPrintf("Candidate %d get a vote, %d votes so far", rf.me, grantedVotes)
				if grantedVotes * 2 > len(rf.peers) {

					rf.mu.Lock()
					rf.status = 2
					rf.mu.Unlock()

					return
				}


			}else if reply.Term > rf.currentTerm {
				DPrintf("Candidate %d outdated, converted to follower", rf.me)

				rf.mu.Lock()
				rf.status = 0

				rf.mu.Unlock()
				return
			}else{
				DPrintf("Candidate %d did not get a vote", rf.me)
			}
			case <- time.After(electionTimeout):{
				DPrintf("Candidate %d has election timeout at term %d", rf.me, curTerm)
				return //election timeout - start new election!
			}
		}
	}
}

func (rf *Raft) RunLeader() {
	var curTerm int
	var curStatus int

	for{
		DPrintf("Leader: %d at term %d trying to grab locks" , rf.me, curTerm)
		rf.mu.Lock()
		curStatus = rf.status
		curTerm = rf.currentTerm
		rf.mu.Unlock()

		if curStatus == 2 {
			//DPrintf("Leader: %d at term %d sending heartbeats" , rf.me, curTerm)
		}else{
			DPrintf("!!!Leader %d at term %d state out of sync, exit leader state" , rf.me, curTerm)
			return
		}


		for i := 0; i < len(rf.peers); i++  {
			if i == rf.me {
				continue
			}

			go func(k int) {
				args := &AppendEntriesArgs{}
				args.Term = curTerm
				args.LeaderId = rf.me

				reply := &AppendEntriesReply{}
				ok := rf.sendAppendEntries(k, args, reply)

				if ok &&  reply.Term > curTerm {
					DPrintf("Leader: %d at term %d find itself stale during heartbeat, new term at %d" , rf.me, curTerm, reply.Term)

					rf.mu.Lock()
					if (rf.currentTerm  <  reply.Term){
						DPrintf("!!!!!Leader: %d at term %d find itself stale during heartbeat, new term at %d, ready to become follower" , rf.me, curTerm, reply.Term)
						rf.status = 0
						rf.currentTerm = reply.Term
						rf.votedFor = -1
					}
					rf.mu.Unlock()
					return
				}
			}(i)
		}

		time.Sleep(120 * time.Millisecond)

		/*
		if curStatus != 2 {
			DPrintf("stale leader: %d at term %d going to follower" , rf.me, curTerm)
			return
		}
		*/
	}
}

func (rf *Raft) MainLoop() {
	var status int  = 0
	for {
		rf.mu.Lock()
		status = rf.status
		rf.mu.Unlock()

		if status == 0{
			rf.RunFollower()
		}else if status == 1 {
			rf.RunCandidate()
		}else if status == 2 {
			rf.RunLeader()
		}
	}
}


func Make(peers []*labrpc.ClientEnd, me int,
persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	rf.votedFor = -1
	//server starts as follower
	rf.status = 0

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	rf.fHB = make(chan int, len(rf.peers))

	//new server starts as follower
	go func() {
		rf.MainLoop();
	}()

	return rf
}
