# Caster

This is a clustered cache database called Caster (Cache + Cluster).   
The setup and configuration are quite easy.  

I focused mostly on designing a minimal architecture and not overengineering stuff.  
Although cluster management and communication protocol could be better, the architecture can remain the same.  

This project has a cache interface and implements this interface using a LRU cache algorithm.  
Other algorithms can be used such as TTL-based caching, etc.

All `GET`, `SET` and `FLUSH` commands are done in **O(1)** time complexity.  

P.S. this is one of my university projects.
