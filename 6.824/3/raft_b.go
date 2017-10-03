package raft

import (
	"sync"
	"labrpc"
	"time"
	"math/rand"
	"bytes"
	"encoding/gob"
)

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
	CurrentTerm int
	VotedFor int
	status int
	fHB chan int

	Log []LE
	commitIndex int
	lastApplied int
	nextIndex []int
	matchIndex []int
	LastIncludedIndex int

	LastIncludedTerm int
	applyCh chan ApplyMsg
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int
	var isleader bool
	term = rf.CurrentTerm
	isleader = rf.status == 2
	return term, isleader
}

func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.CurrentTerm)
	e.Encode(rf.VotedFor)
	e.Encode(rf.Log)
	e.Encode(rf.LastIncludedIndex)
	e.Encode(rf.LastIncludedTerm)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}


func (rf *Raft) RaftStateSize() int {
	return rf.persister.RaftStateSize()
}

func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		rf.CurrentTerm = 0
		rf.VotedFor = -1
		rf.Log = make([]LE, 0)
		rf.Log = append(rf.Log, LE{nil, 0})
		rf.LastIncludedIndex = -1 
		rf.LastIncludedTerm = -1
		rf.commitIndex = 0
		rf.lastApplied = 0
	}else{
		r := bytes.NewBuffer(data)
		d := gob.NewDecoder(r)
		d.Decode(&rf.CurrentTerm)
		d.Decode(&rf.VotedFor)
		d.Decode(&rf.Log)
		d.Decode(&rf.LastIncludedIndex)
		d.Decode(&rf.LastIncludedTerm)

		if rf.LastIncludedIndex >= 0 { 
			rf.commitIndex = rf.LastIncludedIndex
			rf.lastApplied = rf.LastIncludedIndex
		}else {
			rf.commitIndex = 0
			rf.lastApplied = 0
		}
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

func (rf *Raft) ToPI(vi int) int {
	return vi - rf.LastIncludedIndex - 1 
}


func (rf *Raft) ToVI(pi int) int {
	return pi + rf.LastIncludedIndex +1 
}


func (rf *Raft) CheckOldTermWithinLock(term int) {
	if(term > rf.CurrentTerm){
		rf.status = 0
		rf.CurrentTerm = term
		rf.VotedFor = -1
	}
}

func (rf *Raft) GetTerm(vi int) int {
	if (vi == rf.LastIncludedIndex) {
		return rf.LastIncludedTerm
	}else{
		return rf.Log[rf.ToPI(vi)].Term
	}
}

func (rf *Raft) MoreUpToDateWithinLock(args *RequestVoteArgs) bool {
	nextIndex := rf.ToVI(len(rf.Log))

	lastLETerm := rf.GetTerm(nextIndex -1)

	return lastLETerm > args.LastLogTerm || (lastLETerm == args.LastLogTerm && nextIndex - 1 > args.LastLogIndex) 
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()

	if args.Term < rf.CurrentTerm {
		DPrintf("%d of status %d, term %d refused a vote from %d at term %d", rf.me, rf.status, rf.CurrentTerm, args.CandidateId, args.Term)
		reply.VoteGranted = false
		reply.Term = rf.CurrentTerm
	}else{
		rf.CheckOldTermWithinLock(args.Term)

		if rf.VotedFor < 0 || rf.VotedFor == args.CandidateId  {

			if (rf.MoreUpToDateWithinLock(args)){
				DPrintf("%d of status %d, term %d can not vote for %d at term %d, because its Log is more up-topdate", rf.me, rf.status, rf.CurrentTerm, args.CandidateId, args.Term)
				reply.VoteGranted = false
				reply.Term = rf.CurrentTerm
			}else{
				DPrintf("%d of status %d, term %d grants a vote to %d at term %d", rf.me, rf.status, rf.CurrentTerm, args.CandidateId, args.Term)
				reply.VoteGranted = true
				reply.Term =  args.Term
				rf.VotedFor = args.CandidateId
			}
		}else{
			DPrintf("%d of status %d, term %d refuses to vote for %d at term %d, because it voted already this term", rf.me, rf.status, rf.CurrentTerm, args.CandidateId, args.Term)
			reply.VoteGranted = false
			reply.Term =  args.Term
		}
	}

	rf.mu.Unlock()

	if reply.VoteGranted {
		rf.fHB <- args.Term
	}
}

type InstallSnapshotArgs struct {
	Term int
	LeaderId int
	LastIncludedIndex int
	LastIncludedTerm int
	Data []byte
}

type InstallSnapshotReply struct {
	Term int
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
	ConfI int
}

func (rf *Raft) AppendToLogWithinLock(args *AppendEntriesArgs) {
	newIndex := 0
	newELen := len(args.Entries)

	nextIndex := rf.ToVI(len(rf.Log))

	for ;newIndex < newELen; newIndex++ {

		existingIndex := args.PrevLogIndex + newIndex + 1

		if existingIndex >= nextIndex {
			break
		}

		if rf.GetTerm(existingIndex) != args.Entries[newIndex].Term {
			rf.Log = rf.Log[:rf.ToPI(existingIndex)]
			break
		}
	}

	for ;newIndex < newELen; newIndex++ {
		rf.Log = append(rf.Log, args.Entries[newIndex])
	}

	if args.LeaderCommit > rf.commitIndex {
		DPrintf("%d of status %d, term %d updating commit index currently at %d, leaderCommit:%d, index of last new entry: %d", rf.me, rf.status, rf.CurrentTerm, rf.commitIndex, args.LeaderCommit, args.PrevLogIndex + newELen)

		if args.LeaderCommit < args.PrevLogIndex + newELen {
			rf.commitIndex = args.LeaderCommit
		} else {
			rf.commitIndex = args.PrevLogIndex + newELen 
		}
	}

	if newELen > 0 {
		DPrintf("%d of end status %d, term %d agree to append entries of length %d for server %d at term %d, after apply: commitIndex: %d, lastApplied: %d, last Log index: %d", rf.me, rf.status, rf.CurrentTerm, newELen, args.LeaderId, args.Term, rf.commitIndex, rf.lastApplied, rf.ToVI(len(rf.Log) - 1))
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	nextIndex := rf.ToVI(len(rf.Log))

	if args.Term < rf.CurrentTerm {
		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d, because request is from the past", rf.me, rf.status, rf.CurrentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.CurrentTerm
		reply.ConfI = nextIndex
		rf.mu.Unlock()
		return
	}

	rf.CheckOldTermWithinLock(args.Term)

	//TODO: disjoint index check here!!!

	if args.PrevLogIndex < rf.LastIncludedIndex {

		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d, because PrevLogIndex is BEFORE snapshot", rf.me, rf.status, rf.CurrentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.CurrentTerm
		reply.ConfI = rf.LastIncludedIndex + 1
		rf.persist()
		rf.mu.Unlock()
		return
	}

	if nextIndex <= args.PrevLogIndex  || rf.GetTerm(args.PrevLogIndex) != args.PrevLogTerm {
		DPrintf("%d of status %d, term %d refused to append entries for %d at term %d, because prevLogTerm does not match the local Log", rf.me, rf.status, rf.CurrentTerm, args.LeaderId, args.Term)
		reply.Success = false
		reply.Term = rf.CurrentTerm

		if nextIndex <= args.PrevLogIndex {
			reply.ConfI = nextIndex
		}else {
			badTerm :=  rf.GetTerm(args.PrevLogIndex)
			i := args.PrevLogIndex - 1

			for ; i > rf.LastIncludedIndex; i-- {
				if rf.GetTerm(i) != badTerm {
					break
				}
			}

			if i == rf.LastIncludedIndex {
				DPrintf("!!!%d of status %d, term %d trying to append prevLogIndex at %d with term %d, LastIncludedIndex at %d, with LastIncludedTerm %d", rf.me, rf.status, rf.CurrentTerm, args.LeaderId, args.Term, args.PrevLogIndex, rf.GetTerm(args.PrevLogIndex), rf.LastIncludedIndex, rf.LastIncludedTerm )

			}

			reply.ConfI = i + 1
		}

		rf.persist()
		rf.mu.Unlock()
		return
	}

	rf.AppendToLogWithinLock(args)

	reply.Success = true
	reply.Term = args.Term
	reply.ConfI = nextIndex

	if rf.status != 0 {
		DPrintf("%d of status %d, term %d ready to append, and thus becoming follower!", rf.me, rf.status, rf.CurrentTerm)
		rf.status = 0 //this server may be a candidate from an isolated session
	}

	rf.persist()
	rf.mu.Unlock()
	//we don't want to block while holding the lock!
	rf.fHB <- args.Term

	//DPrintf("%d of status %d, term %d agree to append entries for %d at term %d, commitIndex: %d, lastApplied: %d RETURNING!", rf.me, rf.status, rf.CurrentTerm, args.LeaderId, args.Term, rf.commitIndex, rf.lastApplied)
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
	defer rf.persist()

	index := -1
	term := -1
	isLeader := rf.status == 2

	if isLeader {
		index = rf.ToVI(len(rf.Log))
		term = rf.CurrentTerm
		rf.Log = append(rf.Log, LE{command, term})
		rf.nextIndex[rf.me] = index + 1
		rf.matchIndex[rf.me] = index

		DPrintf("START Leader %d at term %d just appended locally, local next index at %d", rf.me, rf.CurrentTerm, rf.nextIndex[rf.me])
	}

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	// Your code here, if desired.
}

func randomElectionTimeout() time.Duration {
	return time.Duration(rand.Intn(400) + 400) * time.Millisecond
}

func (rf *Raft) RunFollower() {
	DPrintf("Follower: %d" , rf.me)

	electionTimeout := randomElectionTimeout()

	for {
		select {
		case <- rf.fHB:

			//DPrintf("Follower(?) %d got a heartbeat" , rf.me)
			case <- time.After(electionTimeout):{
				rf.mu.Lock()
				if rf.status == 0 {
					rf.status = 1
				}
				rf.mu.Unlock()
				return
			}
		}
	}
}

func (rf *Raft) CollectRV(cRVR chan RequestVoteReply, curTerm int) int {
	var grantedVotes int = 1

	electionTimeout := randomElectionTimeout()
	for {
		select {
		case reply := <- cRVR:
			if reply.VoteGranted {

				grantedVotes++
				//DPrintf("Candidate %d get a vote, %d votes so far", rf.me, grantedVotes)
				if grantedVotes * 2 > len(rf.peers) {
					DPrintf("Candidate %d get a vote, %d votes so far", rf.me, grantedVotes)
					return 2
				}


			}else if reply.Term > curTerm {
				DPrintf("Candidate %d outdated, converted to follower", rf.me)
				return 0
			}else{
				DPrintf("Candidate %d did not get a vote", rf.me)
			}
			case <- time.After(electionTimeout):{
				DPrintf("Candidate %d has election timeout at term %d, start new election!", rf.me, curTerm)
				return 1
			}
		}
	}

	return -1
}

func (rf *Raft) BrRV(cRVR chan RequestVoteReply, curTerm int, lastLogIndex int, lastLogTerm int) {
	for  i:=0; i<len(rf.peers);i++ {
		if i == rf.me {
			continue
		}

		go func(k int) {
			args := &RequestVoteArgs{}
			args.CandidateId = rf.me

			args.Term = curTerm
			args.LastLogIndex = lastLogIndex
			args.LastLogTerm = lastLogTerm


			reply := &RequestVoteReply{ }
			ok := rf.sendRequestVote(k, args,  reply)

			if ok {
				cRVR <- *reply
			}else{
				DPrintf("Candidate %d to %d vote request is not ok, term: %d", rf.me, k, curTerm)
			}
		}(i)
	}


}

func (rf *Raft) RunCandidate() {
	rf.mu.Lock()
	if rf.status != 1 {
		DPrintf("Candidate %d term %d actually at status %d, becoming a follower", rf.me, rf.CurrentTerm, rf.status) 
		rf.status = 0
		rf.mu.Unlock()
		return
	}
	rf.CurrentTerm= rf.CurrentTerm  + 1
	rf.VotedFor = rf.me
	curTerm := rf.CurrentTerm 
	lastLogIndex := rf.ToVI(len(rf.Log) - 1)
	lastLogTerm := rf.GetTerm(lastLogIndex)
	rf.persist()
	rf.mu.Unlock()

	DPrintf("Candidate: %d at term: %d" , rf.me, curTerm)

	cRVR := make(chan RequestVoteReply, len(rf.peers) * 5)

	rf.BrRV(cRVR, curTerm, lastLogIndex, lastLogTerm)
	nextStatus := rf.CollectRV(cRVR, curTerm)

	if nextStatus < 0 {
		DPrintf("THIS SHOULD NOT HAPPEN!")
	}

	rf.mu.Lock()
	if rf.status != 1 {
		DPrintf("Candidate %d at term %d currently has status %d?!", rf.me, rf.CurrentTerm, rf.status)
		rf.status = 0
	}else {
		rf.status = nextStatus
	}
	rf.mu.Unlock()
}


func (rf *Raft) UpdateCIWithinLock() {
	for N := rf.commitIndex + 1; N < rf.ToVI(len(rf.Log)); N++ {

		if N < rf.LastIncludedIndex {
			DPrintf("Server %d trying to update commit index at %d to %d, but it is stale data! snapshot ends at index %d", rf.me, rf.commitIndex, N, rf.LastIncludedIndex)
		}

		var goodC int = 0
		for i := 0; i < len(rf.peers); i++ {
			if rf.matchIndex[i]  >= N && rf.GetTerm(N) == rf.CurrentTerm {
				goodC++
			}
		}

		if goodC * 2 > len(rf.peers) {
			rf.commitIndex = N
			DPrintf("Leader: %d at term %d has new commitIndex %d", rf.me, rf.CurrentTerm, rf.commitIndex)
		}
	}
}

func (rf *Raft) RepToAllFs(){
	for i := 0; i < len(rf.peers); i++  {
		if i == rf.me {
			continue
		}

		go func(k int) {
			rf.InstallSnapshotRPC(k)
			rf.RepToFollower(k)
		}(i)
	}
}


func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()

	reply.Term = rf.CurrentTerm
	if args.Term < rf.CurrentTerm{
		rf.mu.Unlock()
		return
	}

	if args.LastIncludedIndex <= rf.LastIncludedIndex {
		rf.mu.Unlock()
		return
	}

	rf.CheckOldTermWithinLock(args.Term)

	if rf.status != 0 {
		DPrintf("!!!!install snapshot to %d of status %d??!!", rf.me, rf.status)
	}

	lastIndex := rf.ToVI(len(rf.Log) - 1)

	if lastIndex <= args.LastIncludedIndex {
		rf.Log = make([]LE, 0)
	}else{
		keptLen := lastIndex - args.LastIncludedIndex

		keptStartI := lastIndex - keptLen + 1

		rf.Log = rf.Log[rf.ToPI(keptStartI):]
	}


	DPrintf("%d at term %d, status %d installing snapshot from leader %d at term %d, before len: %d, before LII:%d, after len:%d  after LII %d", rf.me, rf.CurrentTerm, rf.status, args.LeaderId, args.Term, rf.ToPI(lastIndex + 1), rf.LastIncludedIndex, len(rf.Log), args.LastIncludedIndex)

	rf.persister.SaveSnapshot(args.Data)
	rf.LastIncludedIndex = args.LastIncludedIndex
	rf.LastIncludedTerm = args.LastIncludedTerm
	rf.lastApplied = args.LastIncludedIndex
	rf.commitIndex = args.LastIncludedIndex

	if (rf.lastApplied > args.LastIncludedIndex || rf.commitIndex > args.LastIncludedIndex){
		DPrintf("%d at term %d, status %d installing snapshot from leader %d at term %d, before len: %d, before LII:%d, after len:%d  after LII %d, rewinding???!!!", rf.me, rf.CurrentTerm, rf.status, args.LeaderId, args.Term, rf.ToPI(lastIndex + 1), rf.LastIncludedIndex, len(rf.Log), args.LastIncludedIndex)
	}

	rf.persist()
	rf.mu.Unlock()
	applyMsg :=  ApplyMsg { args.LastIncludedIndex, nil, true, args.Data}

	rf.fHB <- args.Term
	rf.applyCh <- applyMsg
}

func (rf *Raft) CreateAppendArgWithinLock(k int) AppendEntriesArgs {
	args := AppendEntriesArgs{}
	args.Term = rf.CurrentTerm
	args.LeaderId = rf.me

	lastLogIndex := rf.ToVI(len(rf.Log) - 1)

	nextIndexF := rf.nextIndex[k]
	if lastLogIndex >= nextIndexF {
		DPrintf("RepToFollower: Leader: %d at term %d appending to server %d starting from index %d" , rf.me, args.Term, k, nextIndexF)

		args.PrevLogIndex = nextIndexF - 1
		args.PrevLogTerm = rf.GetTerm(args.PrevLogIndex)
		args.Entries = make([]LE, lastLogIndex  -  nextIndexF + 1) 

		//Deep copy so that rf.Log will not be referred indirectly outside lock by go slices
		copy(args.Entries, rf.Log[rf.ToPI(nextIndexF):])
	} else {
		//DPrintf("Leader: %d at term %d, lastLogIndex %d, appending to %d, nextIndex at %d as heartbeat??? " , rf.me, args.Term, lastLogIndex, k, rf.nextIndex[k])
		args.PrevLogIndex = lastLogIndex
		args.PrevLogTerm = rf.GetTerm(args.PrevLogIndex)
		args.Entries = make([]LE, 0)
	}

	args.LeaderCommit = rf.commitIndex

	return args
}

func (rf *Raft) InstallSnapshotRPC(k int) {
	rf.mu.Lock()
	prevNextI := rf.nextIndex[k]

	//No need to trigger snapshot from leader
	if rf.status != 2 || prevNextI > rf.LastIncludedIndex {
		rf.mu.Unlock()
		return
	}

	args := InstallSnapshotArgs{}
	args.Term = rf.CurrentTerm
	args.LeaderId = rf.me
	args.LastIncludedIndex = rf.LastIncludedIndex
	args.LastIncludedTerm = rf.LastIncludedTerm
	args.Data = rf.persister.snapshot

	rf.mu.Unlock()

	reply := &InstallSnapshotReply{}

	ok := rf.peers[k].Call("Raft.InstallSnapshot", &args, reply)

	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	if ok {
		if reply.Term > rf.CurrentTerm {
			DPrintf("!!!!!Leader: %d at term %d find itself stale during install snapshot, new term at %d, ready to become follower" , rf.me, rf.CurrentTerm, reply.Term)
			rf.status = 0
			rf.CurrentTerm = reply.Term
			rf.VotedFor = -1
		} else {
			if prevNextI == rf.nextIndex[k] {
				rf.nextIndex[k] =  args.LastIncludedIndex + 1
				rf.matchIndex[k] =  args.LastIncludedIndex
			}else{
				DPrintf("!!!Next index changed between install SS calls??")
			}
		}
	}
}

func (rf *Raft) RepToFollower(k int) {
	for {
		rf.mu.Lock()
		if rf.status != 2 || rf.nextIndex[k] <= rf.LastIncludedIndex {
			rf.mu.Unlock()
			return
		}

		lastLogIndex := rf.ToVI(len(rf.Log) - 1)
		args := rf.CreateAppendArgWithinLock(k)

		rf.mu.Unlock()

		reply := &AppendEntriesReply{}
		ok := rf.sendAppendEntries(k, &args, reply)

		if !ok {
			//DPrintf("NOT OK: Leader: %d append to server %d at lastLogindex %d" , rf.me, k, lastLogIndex)
			return
		}

		rf.mu.Lock()
		if rf.status != 2 {
			DPrintf("!!!!!Leader: %d at term %d find itself is actually at status %d, therefore becoming a follower" , rf.me, rf.CurrentTerm, rf.status)
			rf.status = 0
			rf.mu.Unlock()
			return
		} else if reply.Term > rf.CurrentTerm {
			DPrintf("!!!!!Leader: %d at term %d find itself stale during heartbeat, new term at %d, ready to become follower" , rf.me, rf.CurrentTerm, reply.Term)
			rf.status = 0
			rf.CurrentTerm = reply.Term
			rf.VotedFor = -1
			rf.persist()
			rf.mu.Unlock()
			return
		}else if (reply.Success) {
			curLastLogIndex := rf.ToVI(len(rf.Log) - 1)
			if len(args.Entries) > 0 {
				DPrintf("Leader: %d at term %d append to server %d at term %d successfully, update that server's matchIndex to %d, leader's current lastIndex at %d" , rf.me, rf.CurrentTerm, k, reply.Term, lastLogIndex, curLastLogIndex)
			}
			rf.matchIndex[k] = lastLogIndex
			rf.nextIndex[k] = lastLogIndex + 1

			if lastLogIndex + 1 != reply.ConfI {
				DPrintf("Leader: %d at term %d append to server %d at term %d successfully, update that server's matchIndex to %d, leader's current lastIndex at %d, next conflicting index at %d" , rf.me, rf.CurrentTerm, k, reply.Term, lastLogIndex, curLastLogIndex, reply.ConfI)
			}

			rf.UpdateCIWithinLock()
			rf.mu.Unlock()
			return
		}else {
			DPrintf("Leader: %d at term %d UNABLE TO append to server %d at lastLogindex %d, reducing nextIndex currently at %d to %d" , rf.me, rf.CurrentTerm, k, lastLogIndex, rf.nextIndex[k], reply.ConfI)

			rf.nextIndex[k] = reply.ConfI
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) RunLeader() {

	var curTerm int
	var curStatus int

	rf.mu.Lock()
	curStatus = rf.status
	curTerm = rf.CurrentTerm

	if curStatus != 2 {
		DPrintf("!!!Leader %d at term %d state out of sync at start, exit leader state" , rf.me, curTerm)
		rf.mu.Unlock()
		return
	}else{

		for i := 0; i < len(rf.peers); i++  {
			rf.nextIndex[i] = rf.ToVI(len(rf.Log))
			rf.matchIndex[i] = 0 
		}
		rf.mu.Unlock()
	}

	for{
		rf.mu.Lock()
		curStatus = rf.status
		curTerm = rf.CurrentTerm
		rf.mu.Unlock()

		//DPrintf("Leader: %d at term %d trying to grab locks" , rf.me, curTerm)

		if curStatus == 2 {
			//DPrintf("Leader: %d at term %d sending heartbeats" , rf.me, curTerm)
		}else{
			DPrintf("!!!Leader %d at term %d state out of sync, exit leader state" , rf.me, curTerm)
			return
		}

		rf.RepToAllFs()
		time.Sleep(150 * time.Millisecond)
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

func (rf *Raft) UpdateSS(lastIncludedI int, data [] byte) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	curLastI := rf.ToVI(len(rf.Log) - 1)

	if lastIncludedI <= rf.LastIncludedIndex || lastIncludedI > curLastI {
		return
	}

	lastIncludedTerm := rf.GetTerm(lastIncludedI)

	if lastIncludedI == curLastI {
		rf.Log = make([]LE, 0)
	} else {
		rf.Log = rf.Log[rf.ToPI(lastIncludedI + 1):]
	}

	DPrintf("%d at term %d, status %d snapshots from its own kvserver, before len: %d, before LII:%d, after len:%d  after LII %d", rf.me, rf.CurrentTerm, rf.status, curLastI - rf.LastIncludedIndex , rf.LastIncludedIndex, len(rf.Log), lastIncludedI)

	rf.persister.SaveSnapshot(data)
	rf.LastIncludedIndex = lastIncludedI
	rf.LastIncludedTerm = lastIncludedTerm

	if rf.LastIncludedIndex > rf.commitIndex {
		DPrintf("!!!snapshot last index is after commit??")
	}
}


func (rf *Raft) ApplyLog(applyCh chan ApplyMsg) {
	for {
		rf.mu.Lock()
		if rf.commitIndex > rf.lastApplied{
			rf.lastApplied++
			msgToApply := ApplyMsg{rf.lastApplied, rf.Log[rf.ToPI(rf.lastApplied)].Command, false, nil}

			if msgToApply.Command == nil {
				DPrintf("!!!%d at term %d, status %d gives an empty command at index %d", rf.me, rf.CurrentTerm, rf.status, rf.lastApplied)
			}

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
	//server starts as follower
	rf.status = 0

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	rf.fHB = make(chan int, len(rf.peers))
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	rf.applyCh = applyCh
	//new server starts as follower
	go func() {
		rf.MainLoop();
	}()

	go func()  {
		rf.ApplyLog(applyCh);
	}()

	return rf
}
