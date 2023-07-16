# Caster

This is a clustered cache database called Caster (Cache + Cluster).  
The setup and configuration are quite easy.  

I didn't try to choose the best networking desicions because I just wanted to design a minimal architecture and not get too deep.

Cluster management and communication protocol could be better.

All `GET`, `SET` and `FLUSH` commands are done in **O(1)** time complexity.

This project has cache interface and implements this interface using a LRU cache algorithm.
Other algorithms can be used such as TTL based caching, etc.

P.S. this is one of my university projects.
