Model networks as graphs. Endpoints on outskirts, switches (a
 type of "middlebox") in the middle. Edge = direct connection
 between two nodes (perhaps a wire, perhaps not)

Routing: A common protocol is link-state routing. Nodes flood
 the network with advertisements, use those advertisements to
 infer the topology, and run Dijkstra's algorithm on that
 topology. This protocol is run repeatedly to catch failures.

Why did the Internet architects decide to divide TCP and IP into two separate protocols?
How do datagrams help the Internet achieve two of its goals: to connect existing networks and to survive transient failures?

packet switching with circut switching, where the relevant state is kept on switches
instead of attached to teach packet.

Once a switch has run a routing protocol, when it gets a packet
destined for Machine A, it knows which switch to send the packet to next.

 If a packet arrives when the queue is full, that packet will be
dropped; it disappears from the network. Full queues are one reason that a network might
experience loss. Other reasons are if a switch crashes or if a link fails (e.g., an Ethernet cable
gets cut).

Does it provide multicast (broadcast): the ability to send a piece of data to multiple endpoints at once,
instead of just one? 

Working from the top down, an application starts by asking the end-toend
layer to transmit a message or a stream of data to a correspondent. The end-to-end
layer splits long messages and streams into segments, it copes with lost or duplicated segments,
it places arriving segments in proper order, it enforces specific communication
semantics, it performs presentation transformations, and it calls on the network layer to
transmit each segment. The network layer accepts segments from the end-to-end layer,
constructs packets, and transmits those packets across the network, choosing which links
to follow to move a given packet from its origin to its destination. The link layer accepts
packets from the network layer, and constructs and transmits frames across a single link
between two forwarders or between a forwarder and a customer of the network. 


Link layer
----------
The call to the link layer identifies a packet buffer
named pkt and specifies that the link layer should place the packet in a frame suitable for
transmission over link2, the link to packet switch C. Switches B and C both have implementations
of the link layer

The combination of the header, payload, and trailer
becomes the link-layer FRAME. The receiving link layer module will, after establishing that
the frame has been correctly received, remove the link layer header and trailer before
passing the payload to the network layer. 

Network layer
---------
The interface to the
network layer, again somewhat simplified, resembles that of the link layer:
NETWORK_SEND (segment_buffer, network_identifier, destination) 

The NETWORK_SEND procedure transmits the segment found in segment_buffer (the payload,
from the point of view of the network layer), using the network named in
network_identifier (a single computer may participate in more than one network), to destination
(the address within that network that names the network attachment point to
which the segment should be delivered). -> PACKET

Next, the network layer consults its tables to choose the most appropriate link over
which to send this packet with the goal of getting it closer to its destination. Finally, the
network layer calls the link layer asking it to send the packet over the chosen link. 

Transport services. Dividing streams and messages into segments and dealing with
lost, duplicated, and out-of-order segments. For this purpose, the end-to-end
header might contain serial numbers of the segments. 

Session services. Negotiating a search, handshake, and binding sequence to locate
and prepare to use a service that knows how to perform the requested procedure.
For this purpose, the end-to-end header might contain a unique identifier that
tells the service which client application is making this call. 

For example, an application that consists of sending a file to a printer would find most
useful a transport service that guarantees to deliver to the printer a stream of bytes in the
same order in which they were sent, with none missing and none duplicated. But a file
transfer application might not care in what order different blocks of the file are delivered,
so long as they all eventually arrive at the destination. A digital telephone application
would like to see a stream of bits representing successive samples of the sound waveform
delivered in proper order, but here and there a few samples can be missing without interfering
with the intelligibility of the conversation. This rather wide range of application
requirements suggests that any implementation decisions that a lower layer makes (for
example, to wait for out-of-order segments to arrive so that data can be delivered in the
correct order to the next higher layer) may be counterproductive for at least some applications.
Instead, it is likely to be more effective to provide a library of service modules
that can be selected and organized by the programmer of a specific application. 

This argument against additional layers is an example of a design principle known as => The application knows best. 

TODO: read 7.2.6
