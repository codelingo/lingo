# lingo
A CLI tool to review source code

The mock results can be found in lingo/mock/mock.json - results are keyed of
the directory name.

To add a new example, create a new dir under examples with source code and add
the mock result to the above file.

# to test:
cd go-kit/kit
git checkout 95fa413
cd codelingo/lingo
go install

# TODOS!!
- verifyAuth is disabled
- syncing repo local to remote before review is disabled