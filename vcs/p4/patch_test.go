package p4

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type gitSuite struct{}

var _ = Suite(&gitSuite{})

func (s *gitSuite) TestPatch(c *C) {

}
