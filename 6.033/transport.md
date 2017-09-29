- Basics:
 - Every data packet gets a sequence number (1, 2, 3, ...)
 - Sender has W outstanding packets at any given time. W = window
 size
too small → underutilized network
too large → congestion
 - When receive gets a packet, it sends an ACK back. ACKs are
 cumulative: An ACK for X indicates "I have received all packets
 up to and including X."
 - If sender doesn't receive an ACK indicating that packet X has
 been received, after some amount of time it will "timeout" and
 retransmit X.
 - Maybe X was lost, its ACK was lost, or its ACK is delayed
 - The timeout = proportional to (but a bit larger than) the RTT
 of the path between sender and receiver
 - At receiver: keep buffer to avoid delivering out-of-order
 packets, keep track of last-packet-delivered to avoid delivering
 duplicates.

 Congestion Control:Share bandwidth fairly among all connections that are using itj

 - Need a signal for congestion in the network, so senders can react
 to it.
 - Our signal: packet drops
 - Every RTT:
 - If there is no loss, W = W+1
 - If there is loss, W = W/2
 - This is "Additive Increase Multiplicative Decrease" (AIMD)
 
Senders also use SLOW-START and FAST-RETRANSMIT/FASTRECOVERY
to quickly increase the window and recover from
loss. 

- Reasoning: if there is a retransmission due to timeout, then
 there is significant loss in the network, and senders should
 back *way* off.

--------

TCP
 Can result in long delays when routers have too much buffering
 (Bufferbloat)

 doesn't react to congestion until
 queues are full.
 - Full queues = long delay
 - Queues = necessary to absorb bursts
 - Goal: Transient queues, not persistent queues
 - Idea: drop packets *before* the queues are full. TCP senders will
 back off before congestion is too bad.

dropTail

RED/ECN
