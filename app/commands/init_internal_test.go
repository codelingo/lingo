package commands

import (
	"reflect"
	"testing"

	"github.com/codelingo/lingo/vcs"
	"github.com/codelingo/lingo/vcs/mock"
)

func TestRepoNaming(t *testing.T) {
	cases := []struct {
		r                         vcs.Repo
		dirName, expectedRepoName string
		expectedErr               error
	}{
		{
			r:                &mock.Repo{},
			dirName:          "freshPkg",
			expectedRepoName: "freshPkg",
			expectedErr:      nil,
		},
		{
			r:                &mock.Repo{},
			dirName:          "existingPkg",
			expectedRepoName: "existingPkg-1",
			expectedErr:      nil,
		},
		{
			r:                &mock.Repo{},
			dirName:          "existingPkg-1105",
			expectedRepoName: "existingPkg-1106",
			expectedErr:      nil,
		},
		{
			r:                &mock.Repo{},
			dirName:          "existing-Pkg-0",
			expectedRepoName: "existing-Pkg-0-1",
			expectedErr:      nil,
		},
	}

	for _, c := range cases {
		gotName, gotErr := vcs.CreateRepo(c.r, c.dirName)
		if gotName != c.expectedRepoName {
			t.Error(`want: `, c.expectedRepoName, `
		got: `, gotName)
		}
		if reflect.DeepEqual(gotErr, c.expectedErr) == false {
			t.Error(`want: `, c.expectedErr.Error(), `
		got: `, gotErr.Error())
		}
	}

}
