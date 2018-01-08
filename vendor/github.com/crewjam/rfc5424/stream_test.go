package rfc5424

import (
	"bytes"

	. "gopkg.in/check.v1"
)

var _ = Suite(&StreamTest{})

type StreamTest struct {
}

func (s *StreamTest) TestCanReadAndWrite(c *C) {
	stream := bytes.Buffer{}
	for i := 0; i < 4; i++ {
		m := Message{Priority: Priority(i), Timestamp: T("0000-12-31T00:00:00Z")}
		nbytes, err := m.WriteTo(&stream)
		c.Assert(err, IsNil)
		c.Assert(nbytes, Equals, int64(38))
	}

	c.Assert(string(stream.Bytes()), Equals,
		`35 <0>1 0000-12-31T00:00:00Z - - - - -`+
			`35 <1>1 0000-12-31T00:00:00Z - - - - -`+
			`35 <2>1 0000-12-31T00:00:00Z - - - - -`+
			`35 <3>1 0000-12-31T00:00:00Z - - - - -`)

	for i := 0; i < 4; i++ {
		m := Message{Priority: Priority(i << 3)}
		nbytes, err := m.ReadFrom(&stream)
		c.Assert(err, IsNil)
		c.Assert(nbytes, Equals, int64(38))
		c.Assert(m, DeepEquals, Message{Priority: Priority(i),
			Timestamp:      T("0000-12-31T00:00:00Z"),
			StructuredData: []StructuredData{}})
	}
}

func (s *StreamTest) TestRejectsInvalidStream(c *C) {
	stream := bytes.NewBufferString(`99 <0>1 0000-12-31T00:00:00Z - - - - -`)
	for i := 0; i < 4; i++ {
		m := Message{Priority: Priority(i << 3)}
		_, err := m.ReadFrom(stream)
		c.Assert(err, Not(IsNil))
	}
}

func (s *StreamTest) TestRejectsInvalidStream2(c *C) {
	stream := bytes.NewBufferString(`0 <0>1 0000-12-31T00:00:00Z - - - - -`)
	for i := 0; i < 4; i++ {
		m := Message{Priority: Priority(i << 3)}
		_, err := m.ReadFrom(stream)
		c.Assert(err, Not(IsNil))
	}
}
