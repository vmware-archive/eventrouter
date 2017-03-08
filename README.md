# Eventrouter

This repository contains a simple event router for the [Kubernetes][kubernetes] project.  
The event router serves as an active watcher of _event_ resource in the kubernetes system, 
which takes those events and _pushes_ them to a user specified _sink_.  This is useful
for a number of different purposes, but most notably long term behavioral analysis of your 
workloads running on your kubernetes cluster.   

## Goals

This project has several objectives, which include: 

* Persist events for longer period of time to allow for system debugging
* Allows operators to forward events to other system(s) for archiving/ML/introspection/etc. 
* It should be relatively low overhead
* Support for multiple _sinks_ should be configurable

Non-goals: 

* This service does not provide a querable extension, that is a responsibility of the 
_sink_
* This service does not serve as a storage layer, that is also the responsibility of the _sink_

## Building 

You can build and test either as a standalone go application or as a Docker container.

### Standalone:
```
$ go get github.com/heptio/eventrouter
```  

### Containerized: 
```
$ sudo make all 
```

## Configure and Execute


### Standalone:
Assuming your testing on a local cluster, it will use the local `config.json`, which 
you can override if 
```
$ ./eventrouter -v 3 -logtostderr 
```

### On the cluster 
Standup: 
```
$ kubectl create -f yaml/
```
Teardown: 
```
$ kubectl delete -f yaml/
```

[kubernetes]: https://github.com/kubernetes/kubernetes/ "Kubernetes"
