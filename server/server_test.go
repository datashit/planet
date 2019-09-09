package main

import (
	"flag"
	"os"
	"testing"

	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}
func TestValidSequence(t *testing.T) {

	r := validSequence(0, 1)

	assert.OK(t, r == true, "valid sequence")

	r = validSequence(200, 1)

	assert.OK(t, r == false, "valid sequence")

	r = validSequence(33500, 0)

	assert.OK(t, r == true, "valid sequence")
}
