package mock_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/codelingo/demo/lingo/mock"
	jc "github.com/juju/testing/checkers"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type mockSuite struct{}

var _ = Suite(&mockSuite{})

func (s *mockSuite) TestDump(c *C) {
	r, err := mock.Results()
	c.Assert(err, jc.ErrorIsNil)
	spew.Dump(r)
}
