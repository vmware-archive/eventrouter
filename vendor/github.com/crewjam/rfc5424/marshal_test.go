package rfc5424

import (
	"fmt"
	"time"

	. "gopkg.in/check.v1"
)

var _ = Suite(&MarshalTest{})

type MarshalTest struct {
}

func T(s string) time.Time {
	rv, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return rv
}

var testCases = []struct {
	in       Message
	expected string
}{
	// RFC-5424 Example 1
	{Message{
		Priority:       34,
		Timestamp:      T("2003-10-11T22:14:15.003Z"),
		Hostname:       "mymachine.example.com",
		AppName:        "su",
		MessageID:      "ID47",
		StructuredData: []StructuredData{},
		Message:        []byte("'su root' failed for lonvick on /dev/pts/8"),
	}, `<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - 'su root' failed for lonvick on /dev/pts/8`},

	// RFC-5424 Example 2
	{Message{
		Priority:       165,
		Timestamp:      T("2003-08-24T05:14:15.000003-07:00"),
		Hostname:       "192.0.2.1",
		AppName:        "myproc",
		ProcessID:      "8710",
		StructuredData: []StructuredData{},
		Message:        []byte("%% It's time to make the do-nuts."),
	}, `<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.`},

	// RFC-5424 Example 3
	{Message{
		Priority:  165,
		Timestamp: T("2003-10-11T22:14:15.003Z"),
		Hostname:  "mymachine.example.com",
		AppName:   "evntslog",
		MessageID: "ID47",
		StructuredData: []StructuredData{
			StructuredData{
				ID: "exampleSDID@32473",
				Parameters: []SDParam{
					SDParam{
						Name:  "iut",
						Value: "3",
					},
					SDParam{
						Name:  "eventSource",
						Value: "Application",
					},
					SDParam{
						Name:  "eventID",
						Value: "1011",
					},
				},
			},
		},
		Message: []byte("An application event log entry..."),
	}, `<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] An application event log entry...`},

	// RFC-5424 Example 4
	{Message{
		Priority:  165,
		Timestamp: T("2003-10-11T22:14:15.003Z"),
		Hostname:  "mymachine.example.com",
		AppName:   "evntslog",
		MessageID: "ID47",
		StructuredData: []StructuredData{
			StructuredData{
				ID: "exampleSDID@32473",
				Parameters: []SDParam{
					SDParam{
						Name:  "iut",
						Value: "3",
					},
					SDParam{
						Name:  "eventSource",
						Value: "Application",
					},
					SDParam{
						Name:  "eventID",
						Value: "1011",
					},
				},
			},
			StructuredData{
				ID: "examplePriority@32473",
				Parameters: []SDParam{
					SDParam{
						Name:  "class",
						Value: "high",
					},
				},
			},
		},
	}, `<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"][examplePriority@32473 class="high"]`},

	{Message{
		Timestamp: T("0000-12-31T00:00:00Z"),
		StructuredData: []StructuredData{
			StructuredData{
				ID: "x@1",
				Parameters: []SDParam{
					SDParam{
						Name:  "class",
						Value: `backslash=\ quote=" right bracket=] left bracket=[`,
					},
				},
			},
		},
	}, `<0>1 0000-12-31T00:00:00Z - - - - [x@1 class="backslash=\\ quote=\" right bracket=\] left bracket=["]`},

	{Message{
		Timestamp:      T("0000-12-31T00:00:00Z"),
		StructuredData: []StructuredData{},
	}, `<0>1 0000-12-31T00:00:00Z - - - - -`},

	{Message{
		Timestamp: T("0000-12-31T00:00:00Z"),
		StructuredData: []StructuredData{
			StructuredData{
				ID: "x@1",
				Parameters: []SDParam{
					SDParam{
						Name:  "",
						Value: "value",
					},
				},
			},
		},
	}, `<0>1 0000-12-31T00:00:00Z - - - - [x@1 ="value"]`},
}

func (s *MarshalTest) TestCanMarshalAndUnmarshal(c *C) {
	for _, tt := range testCases {
		actual, err := tt.in.MarshalBinary()
		c.Assert(err, IsNil)
		c.Assert(string(actual), Equals, tt.expected)

		m := Message{}
		err = m.UnmarshalBinary(actual)
		if err != nil {
			c.Logf(": %s", actual)
			c.Logf(": %#v", m)
		}
		c.Assert(err, IsNil)
		c.Assert(m, DeepEquals, tt.in)
	}
}

// These two strings form the basis of the invalidStrings below. (We change to
// make sure they are valid to we know our tests are sensitive the way we want
// them to be.
var validStrings = [][]byte{
	[]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su X ID47 - msg`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="value"]`),
}

var invalidStrings = [][]byte{
	[]byte(``),
	[]byte(`<`),
	[]byte(`<3`),
	[]byte(`<34>`),
	[]byte(`<34>1`),
	[]byte(`<34>1 `),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z`),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z `),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com`),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su`),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su X`),
	[]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su X ID47`),
	[]byte(`<F>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - msg`),
	[]byte(`<34>X 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - msg`),
	[]byte(`<34>1 notATimestamp mymachine.example.com su - ID47 - 'su root' failed for lonvick on /dev/pts/8`),
	[]byte(`>34<1 2003-10-11T22:14:15.003Z mymachine.example.com su X ID47 - msg`),
	[]byte(`<3499999999999999999999999999999999>1 2003-10-11T22:14:15.003Z mymachine.example.com su X ID47 - msg`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 `),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 ]`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name=`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="]`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="value`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="value"`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="value"x]`),
	[]byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [id name="value\`),
}

func (s *MarshalTest) TestCannotUnmarshalBrokenStrings(c *C) {
	for _, actual := range validStrings {
		m := Message{}
		err := m.UnmarshalBinary(actual)
		c.Assert(err, IsNil)
	}

	for _, actual := range invalidStrings {
		m := Message{}
		err := m.UnmarshalBinary(actual)
		if err == nil {
			c.Logf(": %s", actual)
			c.Logf(": %#v", m)
		}
		c.Assert(err, Not(IsNil))
		c.Assert(fmt.Sprintf("%s", err), Not(Equals), "")
	}
}

var invalidMessages = []Message{
	Message{Hostname: "\x7f"},
	Message{Hostname: "\x20"},
	Message{Hostname: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAA",
	},
	Message{AppName: "\x7f"},
	Message{AppName: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	Message{ProcessID: "\x7f"},
	Message{ProcessID: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" +
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	},
	Message{MessageID: "\x7f"},
	Message{MessageID: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         "\x20",
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         "\x7f",
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         "foo=bar",
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         "foo[bar]",
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         `foo"bar`,
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         `x@1`,
				Parameters: []SDParam{SDParam{Name: "\x7f", Value: "value"}},
			},
		},
	},
	Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         `x@1`,
				Parameters: []SDParam{SDParam{Name: "x", Value: "\xc3\x28"}},
			},
		},
	},
}

func (s *MarshalTest) TestCannotMarshalInvalidMessages(c *C) {
	for i, m := range invalidMessages {
		bin, err := m.MarshalBinary()
		if err == nil {
			c.Logf(": %d", i)
			c.Logf(": %s", string(bin))
			c.Logf(": %#v", m)
		}
		c.Assert(err, Not(IsNil))
		c.Assert(fmt.Sprintf("%s", err), Not(Equals), "")
	}
}

func (s *MarshalTest) TestLongAttributes(c *C) {

	m := Message{
		StructuredData: []StructuredData{
			StructuredData{
				ID:         "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
				Parameters: []SDParam{SDParam{Name: "", Value: "value"}},
			},
		},
	}
	bin, err := m.MarshalBinary()
	if allowLongSdNames {
		c.Assert(err, IsNil)
		c.Assert(string(bin), Equals, "<0>1 0001-01-01T00:00:00Z - - - - [AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA =\"value\"]")
	} else {
		c.Assert(err, Not(IsNil))
		c.Assert(fmt.Sprintf("%s", err), Not(Equals), "")
	}
}
