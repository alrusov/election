# election

[![Go](https://img.shields.io/badge/go-1.20+-blue.svg)]()
[![License](https://img.shields.io/badge/license-MIT-green.svg)]()
[![Redis](https://img.shields.io/badge/redis-supported-red.svg)]()

Leader election library for Go using Redis (Standalone / Sentinel / Cluster).

The library allows multiple equal nodes to elect a master for independent resources with:
heartbeat, TTL-based failover, fencing token (split-brain protection), forced re-election, graceful resign, jittered heartbeat and callbacks.

Designed for clusters of *1 or more nodes**.

---

## Table of Contents

- [Features](#features)
- [Core Idea](#core-idea)
- [Split-brain Protection (Fencing Token)](#split-brain-protection-fencing-token)
- [Configuration](#configuration)
- [Usage](#usage)
- [Force Re-election](#force-re-election)
- [Graceful Shutdown](#graceful-shutdown)
- [Redis Setup](#redis-setup)
- [Failure Scenarios](#failure-scenarios)
- [Guarantees](#guarantees)
- [Notes](#notes)
- [License](#license)

---

## Features

- ✅ Multi-resource leader election
- ✅ Redis standalone / Sentinel / Cluster support
- ✅ Fencing token (split-brain protection)
- ✅ Heartbeat with configurable TTL
- ✅ Jitter (±10%) to avoid thundering herd
- ✅ Force re-election (admin action)
- ✅ Graceful resign on shutdown
- ✅ Callbacks (OnElected / OnRevoked)
- ✅ Storage abstraction (Redis implementation provided)
- ✅ Production-safe (atomic Lua scripts)
- ✅ Unit tested

---

## Core Idea

For each resource the library stores two Redis keys:

election:{resource}:lock   → nodeID (with TTL)  
election:{resource}:token  → monotonically increasing integer  

Leader election is performed atomically using a Lua script:

- if no master exists → acquire lock and increment token  
- if current node is master → renew TTL  
- otherwise → stay follower  

Only the node with the **highest fencing token** is considered a valid master.

---

## Split-brain Protection (Fencing Token)

In case of Redis Sentinel / Cluster failover or network partition, two nodes may temporarily think they are master.

Fencing token guarantees correctness.

Example:

NodeA (old master) token = 5  
NodeB (new master) token = 6  

Only NodeB is allowed to control the resource because it has the highest token.

Applications must treat the fencing token as authoritative:
any operation on the protected resource must be associated with the current token.

This prevents split-brain even during Redis failover or network partitions.

---

## Configuration

Recommended defaults:

HeartbeatInterval = 2s  
TTL               = 6s  
JitterPercent     = 10  

Invariant:

TTL >= 3 × HeartbeatInterval

Jitter is applied to heartbeat interval (±10%) to avoid synchronized elections.

---

## Usage

### Start election:

```go
mgr.Start(ctx, "resource1")
```

### Check master:

```
if mgr.IsMaster("resource1") {
    // do work
}
```

### With callbacks:

```
Callbacks{
    OnElected: func(resource string, token int64) {
        log.Println("I am master for", resource, "token", token)
    },
    OnRevoked: func(resource string) {
        log.Println("Lost master for", resource)
    },
}
```

### Force Re-election

Administrative operation:

```
mgr.ForceReelection("resource1")
```

Deletes the current lock and triggers a new election.

The fencing token is NOT reset and keeps increasing monotonically.

### Graceful Shutdown

```
mgr.Resign("resource1")
```

## Redis Setup

### Standalone Redis

```
redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
```

### Redis Sentinel

```
redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName: "mymaster",
    SentinelAddrs: []string{"host1:26379", "host2:26379"},
})
```

### Redis Cluster

```
redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{"host1:6379", "host2:6379"},
})
```

The election package works transparently with all modes.

## Failure scenarios

Scenario | Result  
---------|-------
Master crashes | New master elected after TTL  
Redis restart | Election restarts  
Sentinel failover | New master elected, fencing token protects resource  
Network partition | Node with highest token wins  
Force re-election | Immediate new master  
20 nodes start simultaneously | Only one elected  

---

## Guarantees

- ✅ No split-brain (fencing token)
- ✅ Atomic election (Lua)
- ✅ At most one valid master at a time
- ✅ Safe with Redis Sentinel and Cluster
- ✅ Independent election per resource
- ✅ Production ready

---

## Notes

- The fencing token must be used by the application when accessing critical resources.
- Redis persistence should be enabled for token durability.
- Logging and metrics are recommended for production use.

---

## License

MIT
