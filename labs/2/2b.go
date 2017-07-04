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

type LE struct {
	Command interface{}
	Term    int
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

	log []LE
	commitIndex int
	lastApplied int
	nextIndex []int
	matchIndex []int
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
	LastLogIndex int
	LastLogTerm int
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


func (rf *Raft) MoreUpToDateWithinLock(args *RequestVoteArgs) bool {
	curLogLen := len(rf.log)

	lastLETerm := rf.log[curLogLen -1].Term

	return lastLETerm > args.LastLogTerm || (lastLETerm == args.LastLogTerm && curLogLen - 1 > args.LastLogIndex) 
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

	if rf.votedFor < 0 || rf.votedFor == args.CandidateId  { //TODO: log check

		if (rf.MoreUpToDateWithinLock(args)){
			DPrintf("%d of status %d, term %d can not vote for %d at term %d, because its log is more up-topdate", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)

			reply.VoteGranted = false
			reply.Term = rf.currentTerm
		}else{
			DPrintf("%d of status %d, term %d grants a vote to %d at term %d", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)
			reply.VoteGranted = true
			reply.Term =  args.Term
			rf.votedFor = args.CandidateId

			if rf.status == 0{
				rf.fHB <- args.Term
			}
		}
	}else{
		DPrintf("%d of status %d, term %d refuses to vote for %d at term %d, because it voted already this term", rf.me, rf.status, rf.currentTerm, args.CandidateId, args.Term)
		reply.VoteGranted = false
		reply.Term =  args.Term
	}
}

type AppendEntriesArgs struct {
	Term int
	LeaderId  int
	PrevLogIndex int
	PrevLogTerm int 
	Entries [] LE
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term int
	Success bool 
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d, because request is from the past", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}

	rf.CheckOldTermWithinLock(args.Term)

	curLogLen := len(rf.log)

	if curLogLen <= args.PrevLogIndex  || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d, because prevLogTerm does not match the local log", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.currentTerm
		return
	}

	newIndex := 0
	newELen := len(args.Entries)

	for ;newIndex < newELen; newIndex++ {

		existingIndex := args.PrevLogIndex + newIndex + 1

		if existingIndex >= curLogLen {
			break
		}

		if rf.log[existingIndex].Term != args.Entries[newIndex].Term {
			rf.log = rf.log[:existingIndex]
			break
		}
	}

	for ;newIndex < newELen; newIndex++ {
		rf.log = append(rf.log, args.Entries[newIndex])
	}


	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit < args.PrevLogIndex + newELen {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = args.PrevLogIndex + newELen 
		}
	}

	reply.Success = true
	reply.Term = args.Term

	DPrintf("%d of end status %d, term %d agree to append entries for %d at term %d, commitIndex: %d, lastApplied: %d", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term, rf.commitIndex, rf.lastApplied)

	if(rf.status == 0){
		rf.fHB <- args.Term
	}else{
		DPrintf("!!!BAD STATE")
	}

	//DPrintf("%d of status %d, term %d agree to append entries for %d at term %d, commitIndex: %d, lastApplied: %d RETURNING!", rf.me, rf.status, rf.currentTerm, args.LeaderId, args.Term, rf.commitIndex, rf.lastApplied)
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	index := -1
	term := -1
	isLeader := rf.status == 2

	if isLeader {
		index = len(rf.log)
		term = rf.currentTerm
		rf.log = append(rf.log, LE{command, term})
		rf.nextIndex[rf.me] = len(rf.log)
	}

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
	rf.mu.Lock()
	rf.currentTerm= rf.currentTerm  + 1
	rf.votedFor = rf.me
	rf.mu.Unlock()

	//DPrintf("Candidate: %d at term: %d" , rf.me, curTerm)

	cRVR := make(chan RequestVoteReply, len(rf.peers))
	for  i:=0; i<len(rf.peers);i++ {
		if i == rf.me {
			continue
		}

		go func(k int) {
			args := &RequestVoteArgs{}
			args.CandidateId = rf.me

			rf.mu.Lock()
			args.Term = rf.currentTerm
			args.LastLogIndex = len(rf.log) - 1
			args.LastLogTerm = rf.log[args.LastLogIndex].Term
			rf.mu.Unlock()

			reply := &RequestVoteReply{ }
			ok := rf.sendRequestVote(k, args,  reply)

			if ok {
				cRVR <- *reply
			}else{
				//	DPrintf("Candidate %d to %d vote request is not ok, term: %d", rf.me, k, curTerm)
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
				//DPrintf("Candidate %d has election timeout at term %d", rf.me, curTerm)
				return //election timeout - start new election!
			}
		}
	}
}


func (rf *Raft) UpdateCIWithinLock() {
	for N := rf.commitIndex + 1; N < len(rf.log); N++ {

		var goodC int = 0
		for i := 0; i < len(rf.peers); i++ {
			if rf.matchIndex[i]  >= N && rf.log[N].Term == rf.currentTerm {
				goodC++
			}
		}

		if goodC * 2 > len(rf.peers) {
			rf.commitIndex = N
		}else{
			break
		}
	}
}

func (rf *Raft) RepToFollower(k int) {
	for {
		args := &AppendEntriesArgs{}
		args.LeaderId = rf.me

		rf.mu.Lock()
		args.Term = rf.currentTerm

		lastLogIndex := len(rf.log) - 1

		if lastLogIndex >= rf.nextIndex[k]{
			DPrintf("Leader: %d at term %d appending to %d from %d " , rf.me, args.Term, k, rf.nextIndex[k])
			args.Entries = rf.log[rf.nextIndex[k]:]
			args.PrevLogIndex = rf.nextIndex[k] - 1
			args.PrevLogTerm = rf.log[args.PrevLogIndex].Term

		}else  {
			DPrintf("Leader: %d at term %d, lastLogIndex %d, appending to %d, nextIndex at %d as heartbeat??? " , rf.me, args.Term, lastLogIndex, k, rf.nextIndex[k])
			args.PrevLogIndex = lastLogIndex
			args.PrevLogTerm = rf.log[lastLogIndex].Term
			args.Entries = make([]LE, 0) 
		}

		rf.mu.Unlock()

		reply := &AppendEntriesReply{}
		ok := rf.sendAppendEntries(k, args, reply)

		if !ok {
			DPrintf("NOT OK: Leader: %d append to %d at lastLogindex %d" , rf.me, k, lastLogIndex)
			return
		}

		rf.mu.Lock()
		if reply.Term > rf.currentTerm || rf.status == 0 {
			DPrintf("!!!!!Leader: %d at term %d find itself stale during heartbeat, new term at %d, ready to become follower" , rf.me, rf.currentTerm, reply.Term)
			rf.status = 0
			rf.currentTerm = reply.Term
			rf.votedFor = -1
			rf.mu.Unlock()
			return
		}else if (reply.Success) {
			DPrintf("Leader: %d at term %d append to %d at term %d successfully at lastLogindex %d" , rf.me, rf.currentTerm, k, reply.Term, lastLogIndex)
			rf.matchIndex[k] = lastLogIndex
			rf.nextIndex[k] = lastLogIndex + 1
			rf.UpdateCIWithinLock()
			rf.mu.Unlock()
			return;
		}else {
			DPrintf("Leader: %d at term %d UNABLE TO append to %d at last index %d, reducing nextIndex currently at %d " , rf.me, rf.currentTerm, k, lastLogIndex, rf.nextIndex[k])
			rf.nextIndex[k]--
		}

		rf.mu.Unlock()

	}
}

func (rf *Raft) RunLeader() {

	var curTerm int
	var curStatus int

	rf.mu.Lock()
	for i := 0; i < len(rf.peers); i++  {
		rf.nextIndex[i] = len(rf.log)
		rf.matchIndex[i] = 0 
	}
	rf.mu.Unlock()

	for{
		rf.mu.Lock()
		curStatus = rf.status
		curTerm = rf.currentTerm
		rf.mu.Unlock()

		//DPrintf("Leader: %d at term %d trying to grab locks" , rf.me, curTerm)

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
				rf.RepToFollower(k)
			}(i)
		}

		time.Sleep(120 * time.Millisecond)
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


func (rf *Raft) ApplyLog(applyCh chan ApplyMsg) {
	for {
		rf.mu.Lock()
		if rf.commitIndex > rf.lastApplied{
			rf.lastApplied++
			msgToApply := ApplyMsg{rf.lastApplied, rf.log[rf.lastApplied].Command, false, nil}
			rf.mu.Unlock()

			applyCh <- msgToApply
		}else{
			rf.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
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
	rf.commitIndex = 0 
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.log = make([]LE, 0)
	rf.log = append(rf.log, LE{nil, 0})

	//new server starts as follower
	go func() {
		rf.MainLoop();
	}()

	/*
	go func()  {
		rf.ApplyLog(applyCh);
	}()
	*/

	return rf
}
