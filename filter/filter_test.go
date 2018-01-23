package filter

import (
	"testing"

	"k8s.io/api/core/v1"
)

type passesFilterTest struct {
	filter   EventIncludeFilter
	input    *v1.Event
	expected bool
}

var includeAllFilter = EventIncludeFilter{}

var excludeAllFilter = EventIncludeFilter{
	Kind:      []string{"-"},
	Name:      []string{"-"},
	Namespace: []string{"-"},
}

var podOnlyFilter = EventIncludeFilter{
	Kind: []string{"Pod"},
}

var podInDefaultFilter = EventIncludeFilter{
	Kind:      []string{"Pod"},
	Namespace: []string{"default"},
}

var podOrDeploymentInDefaultFilter = EventIncludeFilter{
	Kind:      []string{"Pod", "Deployment"},
	Namespace: []string{"default"},
}

func buildEvent(kind, name, namespace string) *v1.Event {
	return &v1.Event{
		InvolvedObject: v1.ObjectReference{
			Kind:      kind,
			Name:      name,
			Namespace: namespace,
		},
	}
}

var includeAllTests = []passesFilterTest{
	{
		filter:   includeAllFilter,
		input:    buildEvent("AnyKind", "AnyName", "AnyNamespace"),
		expected: true,
	},
}

var excludeAllTests = []passesFilterTest{
	{
		filter:   excludeAllFilter,
		input:    buildEvent("AnyKind", "AnyName", "AnyNamespace"),
		expected: false,
	},
}

var miscTests = []passesFilterTest{
	// Filter in Pods
	{
		filter:   podOnlyFilter,
		input:    buildEvent("Pod", "", ""),
		expected: true,
	},
	{
		filter:   podOnlyFilter,
		input:    buildEvent("Pod", "AnyName", ""),
		expected: true,
	},
	{
		filter:   podOnlyFilter,
		input:    buildEvent("Pod", "", "AnyNamespace"),
		expected: true,
	},
	{
		filter:   podOnlyFilter,
		input:    buildEvent("NotPod", "", ""),
		expected: false,
	},
	{
		filter:   podOnlyFilter,
		input:    buildEvent("NotPod", "AnyName", "AnyNamespace"),
		expected: false,
	},

	// Filter in Pods in default namespace
	{
		filter:   podInDefaultFilter,
		input:    buildEvent("Pod", "AnyName", "default"),
		expected: true,
	},
	{
		filter:   podInDefaultFilter,
		input:    buildEvent("NotPod", "AnyName", "default"),
		expected: false,
	},
	{
		filter:   podInDefaultFilter,
		input:    buildEvent("NotPod", "AnyName", "notdefault"),
		expected: false,
	},

	// Filter in Pods or Deployments in default namespace
	{
		filter:   podOrDeploymentInDefaultFilter,
		input:    buildEvent("Pod", "AnyName", "default"),
		expected: true,
	},
	{
		filter:   podOrDeploymentInDefaultFilter,
		input:    buildEvent("Deployment", "AnyName", "default"),
		expected: true,
	},
	{
		filter:   podOrDeploymentInDefaultFilter,
		input:    buildEvent("Neither", "AnyName", "default"),
		expected: false,
	},
}

func testPassesFilterIncludeAll(t *testing.T) {
	for i, test := range includeAllTests {
		if res := test.filter.Passes(test.input); res != test.expected {
			t.Errorf("PassesFilterIncludeall test %d: result = %v, expected = %v\n",
				i, res, test.expected)
		}
	}
}

func testPassesFilterExcludeAll(t *testing.T) {
	for i, test := range excludeAllTests {
		if res := test.filter.Passes(test.input); res != test.expected {
			t.Errorf("PassesFilterExcludeAll test %d: result = %v, expected = %v\n",
				i, res, test.expected)
		}
	}
}

func testPassesFilterMisc(t *testing.T) {
	for i, test := range miscTests {
		if res := test.filter.Passes(test.input); res != test.expected {
			t.Errorf("PassesFilterMisc test %d: result = %v, expected = %v\n",
				i, res, test.expected)
		}
	}
}

func TestPassesFilter(t *testing.T) {
	t.Run("IncludeAll", testPassesFilterIncludeAll)
	t.Run("ExcludeAll", testPassesFilterExcludeAll)
	t.Run("Misc", testPassesFilterMisc)
}
