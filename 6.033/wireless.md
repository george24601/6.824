
cal 802.11 scenario: multiple clients trying to send their
 data to an access point.
 - Problem: need to prevent two (or more) clients from sending at
 once.

 - Analogy: if everyone in the room talks at once, I can't tell
 what any of you are saying.
 - Real life: communications are ultimately signals. If you're
 sending on the same channel, and two signals overlap in
 space/time, it's impossible to pull those signals apart.
 - When two clients send at the same time, we say their packets
 have collided. Clients know when collisions happen, but we want
 to prevent them if possible.

- MAC -- Media Access Control -- protocols mitigate collisions
 - In wired networks:
 - Two endpoints on a wire can't send at the same time. It's easy
 to coordinate: A sends, then B sends, then A sends, etc.
 - In practice: modern ethernet provides full duplex
 communications. An ethernet cable contains wires running in
 each direction; there is one set for sending in one direction
 and another for the other. Effectively, two separate channels.

3. MACs don't solve everything
 - Hidden terminals: some clients can't sense each other, will send,
 but their communication collides at the receiver.
 - Exposed terminals: some clients *can* sense each other, and so
 won't send at the same time. However, their communications
 *don't* collide at their respective receivers, so it would
 actually be okay to send.

 the quality of the channel can change over time
 (rapidly, in fact).

 Senders try different rates, pick the one that achieves the
 highest throughput after accounting for packet loss.

- Should we design a specific TCP for wireless networks? Maybe, but
 your traffic usually traverses both wireless and wired networks.
 Hard to reason about those interactions.

- In 802.11:
 - Can move around. Your IP address will (likely) change as you
 move from one AP to another.
 => Have to restart all of your TCP connections
 - If we keep IP addresses the same?
 - Might help, but changes our assumption about what an address
 means (right now: it indicates your attachment point to the
 network, not you yourself)
 - We would still likely experience periods of loss when we switch
 APs, causing TCP to timeout
