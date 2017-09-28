what is the benefit of a recursive query?
What are the benefits of DNS's hierarchical design?
Are there any drawbacks to DNS's design?

Sample question
---------
While browsing the Web, you click on the link www.course6.com. Your computer
asks your Domain Name System (DNS) name server, M, to find an IP address for this domain name.
Which of the following is always true of the name resolution process, assuming that all name servers
are configured correctly and no packets are lost?

1. M must contact one of the root name servers to resolve the domain name. 
False. Caching
2. M must contact one of the name servers for course6.com to resolve the domain name.
False. Caching
3. If M had answered a query for the IP address corresponding to www.course6.com at some time in the past, then it will always correctly respond to the current query without contacting any other name server.
False. Cache expires
4. If M has a valid IP address of a functioning name server for course6.com in its cache, then M will always get a response from that name server without any other name servers being contacted.
True, if the network connection is good
