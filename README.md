# Eventrouter

This repository contains a simple event router for the [Kubernetes][kubernetes] project. The event router serves as an active watcher of _event_ resource in the kubernetes system, which takes those events and _pushes_ them to a user specified _sink_. This is useful for a number of different purposes, but most notably long term behavioral analysis of your
workloads running on your kubernetes cluster.

## Goals

This project has several objectives, which include:

- Persist events for longer period of time to allow for system debugging
- Allows operators to forward events to other system(s) for archiving/ML/introspection/etc.
- It should be relatively low overhead
- Support for multiple _sinks_ should be configurable

### NOTE:

By default, eventrouter is configured to leverage existing EFK stacks by outputting wrapped json object which are easy to index in elastic search.

## Non-Goals:

- This service does not provide a querable extension, that is a responsibility of the
  _sink_
- This service does not serve as a storage layer, that is also the responsibility of the _sink_

## Running Eventrouter

Standup:

```
$ kubectl create -f https://raw.githubusercontent.com/heptiolabs/eventrouter/master/yaml/eventrouter.yaml
```

Teardown:

```
$ kubectl delete -f https://raw.githubusercontent.com/heptiolabs/eventrouter/master/yaml/eventrouter.yaml
```

### Inspecting the output

```
$ kubectl logs -f deployment/eventrouter -n kube-system
```

Watch events roll through the system and hopefully stream into your ES cluster for mining, Hooray!

[kubernetes]: https://github.com/kubernetes/kubernetes/ "Kubernetes"

## Running Eventrouter with Skaffold

1.  Make sure you have [skaffold](https://skaffold.dev/docs/getting-started/#installing-skaffold) and [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) installed
2.  Run `skaffold dev` from the root of this repository
3.  Eventrouter will be compiled and run in your minikube cluster. Any change will cause a recompile & restart of this application.
