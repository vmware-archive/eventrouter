# Eventrouter

This repository contains a event router for the [Kubernetes][kubernetes] project.  
The event router serves as an active watcher of _event_ resource in the kubernetes system, 
which takes those events and _pushes_ them to a user specified _sink_.  

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

## Configuring A Sink 
TODO 

[kubernetes] https://github.com/kubernetes/kubernetes/ "Kubernetes"
