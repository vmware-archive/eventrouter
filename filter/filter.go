package filter

import (
	"k8s.io/api/core/v1"
)

// EventIncludeFilter filters in events that should be included. Values within
// a given field are ORed together, while fields themselves are ANDed together.
// If a given field is empty, then that field allows everything through.
//
// For example, to filter in "Pods or DaemonSets with any Name in Namespace "kube-system":
//    Kind: ["Pod", "DaemonSet"]
//    Name: []
//    Namespace: ["kube-system"]
type EventIncludeFilter struct {
	Kind      []string `mapstructure:"Kind"`
	Name      []string `mapstructure:"Name"`
	Namespace []string `mapstructure:"Namespace"`
}

// existsInSlice returns true if the given string exists in the slice, else false.
func existsInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}

	return false
}

// Passes returns true if the event is allowed through this filter, else false.
func (f EventIncludeFilter) Passes(event *v1.Event) bool {
	if len(f.Kind) > 0 {
		if !existsInSlice(event.InvolvedObject.Kind, f.Kind) {
			return false
		}
	}
	if len(f.Name) > 0 {
		if !existsInSlice(event.InvolvedObject.Name, f.Name) {
			return false
		}
	}
	if len(f.Namespace) > 0 {
		if !existsInSlice(event.InvolvedObject.Namespace, f.Namespace) {
			return false
		}
	}

	return true
}
