package main

import (
	"fmt"
	"os"
	"time"

	"github.com/crewjam/rfc5424"
)

func writeMain() {
	m := rfc5424.Message{
		Priority:  rfc5424.Daemon | rfc5424.Info,
		Timestamp: time.Now(),
		Hostname:  "myhostname",
		AppName:   "someapp",
		Message:   []byte("Hello, World!"),
	}
	m.AddDatum("foo@1234", "Revision", "1.2.3.4")
	m.WriteTo(os.Stdout)
}

func readMain() {
	m := rfc5424.Message{}
	_, err := m.ReadFrom(os.Stdin)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("%#v\n", m)
}

func main() {
	readMain()
}
