# Toy Distributed Key-Value Store

This project is part of my bachelor's degree in the distributed systems course. We chose the projects ourselves and this is what I came up with.

My goal was to learn about distributed systems, key value stores in particular, and the RAFT algorithm. For this I took a look at the etcd architecture and reimplemented a toy subset.

## Course Requirements

From our Cryptpad, in which I specified some of the requirements and scope:

```md
### Distributed Key/Value Store [solo]

Ziele:
- verteilte, redundante in-memory Speicherung von Key/Value via HashMaps auf Cluster
- Orientierung an etcd

Details:
- Implementierung in Go
- Interface als JSON (idk?)
- Minimale DB Features
    - Key/Value als Typ string
	- kein Fokus auf klassische Funktionalität - sandbox für distributed algorithms!
- *automated master election*
- *consensus establishment*
- kein Datenverlust durch Nodeausfaull
```

In our atomic estimation poker session, we estimated that this project will take between 8 and 21 hours to complete.

## Documentation


