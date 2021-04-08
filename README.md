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

In our atomic estimation poker session, we estimated that this project will take between 8 and 21(+) hours to complete.

## Testing Setup

- Tests unfortunately currently depend on one another
- Tests require a specific number of followers (>=2)

## Miscellaneous

### asciinema Recording

The following is a link to the asciinema recording of `make run-dc`:

[![asciicast](https://asciinema.org/a/czpEq0uaNrn7X8yjVasnAo19t.svg)](https://asciinema.org/a/czpEq0uaNrn7X8yjVasnAo19t)

### Lines of Code for this Project

```sh
$ cloc . # without vendor
      29 text files.
      29 unique files.                              
       8 files ignored.

github.com/AlDanial/cloc v 1.82  T=0.04 s (509.5 files/s, 47263.9 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              18            308             94           1502
Markdown                         1             20              0             51
YAML                             1              2              0             25
make                             1              6              1             16
Dockerfile                       1              4              3              9
-------------------------------------------------------------------------------
SUM:                            22            340             98           1603
-------------------------------------------------------------------------------
```

### Amount of time spent on this project

I ran a timer during all of the time I worked on this project, including the presentation and
small breaks with a duration of less than 10 minutes.

In total, I spent **24:32:15** (hh:mm:ss) on it.
