package main

import "testing"

func Test_Gretting(t *testing.T) {
	expected := "Hello, World!"
	actual := greeting()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
