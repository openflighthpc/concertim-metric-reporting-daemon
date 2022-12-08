package main

import "testing"

func Test_Nothing(t *testing.T) {
	expected := "Hello, World!"
	actual := "Hello, World!"
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
