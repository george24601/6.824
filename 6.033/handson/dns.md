1. 18.26.0.122
2. canonical names
3. 1800 sec
4. data resolves 18.17.196.100. data. resolves to nothing
5. The first asks the resolve data.mit.edu., the second asks to resolve just data. . 
6. 
```
dig @a.ROOT-SERVERS.NET lirone.csail.mit.edu  +norecurs
```

7.
```
dig @f.edu-servers.net  lirone.csail.mit.edu  +norecurs
dig @usw2.akam.net  lirone.csail.mit.edu  +norecurs
dig @auth-ns0.csail.mit.edu  lirone.csail.mit.edu  +norecurs
```

8.
a. dig www.dmoz.org +norecurs 
b. No, because it gives only the NS records
c. 15ms
9. 5ms with cache
10. 7ms even with norecursion
