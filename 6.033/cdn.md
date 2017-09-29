BitTorrent


P2P networks create overlays on top of the underlying Internet (so  do CDNs)

- Now imagine two clients, both behind NATs
 A --- N1 ---- N2 --- S

 - Now A doesn't even know S's IP (private IPs aren't routable).
 It also doesn't know N2's IP; it has no way to get that.
 - For Skype: means that A and S can't call each other
 - Skype provides a directory service, so assume we can get N2's
 public IP. When N2 gets packet destined for S, it has no idea
 what to do with it.

- Skype will employ an additional node -- a "supernode" -- P, with a
 public IP, and route A and S's calls through P:

 - Today: Microsoft owns all of the supernodes, making this less of a
 P2P network and more of a hierarchy

What's good for streaming? CDNs!
