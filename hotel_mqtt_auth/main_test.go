package main

import "testing"

func TestCheckTopic(t *testing.T)  {
	type testCase struct {
		pattern string
		topic string
		uuid string
		output bool
	}
	testMap := []*testCase{
		{"hotels/%u/room/open", "hotels/a/room/open", "a", true},
		{"hotels/%u/room/open", "hotels/b/room/open", "a", false},
		{"hotels/%u/room/open", "hotels/b/rooms/open", "a", false},
		{"hotels/%u/room/open", "hotels/a/room/open/a", "a", false},
		{"hotels/%u/#/open", "hotels/a/room/open", "a", true},
		{"hotels/%u/#/open", "hotels/a/room/bla/open", "a", true},
		{"hotels/%u/#/open", "hotels/a/room/bla/open/bla", "a", false},
		{"hotels/%u/+/open", "hotels/a/room/bla/open", "a", false},
		{"hotels/%u/+/open/#", "hotels/a/room/open/bla/bla", "a", true},
	}
	for _, test := range testMap {
		output := checkTopic(test.pattern, test.topic, test.uuid)
		if output != test.output {
			t.Fatalf("checkTopic(\"%s\", \"%s\", \"%s\") expected %t, got %t", test.pattern, test.topic,
				test.uuid, test.output, output)
		}
	}
}
