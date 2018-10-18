package main

import "testing"

func TestConntrack(t *testing.T) {
	readConntrack(".samples/ip_conntrack")
}
