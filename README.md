# CodeLingo.io - Code Quality that Scales

## Install

[Download](https://github.com/codelingo/lingo/releases) a pre-built binary or, if you have [Golang setup](https://golang.org/doc/install), install from source:

```bash
$ go get github.com/codelingo/lingo

cd  $GOPATH/src/github.com/go-kit/kit
git checkout 95fa413
./update_deps.bash
```

This will download, build and and place the `lingo` binary on your $PATH

### Windows

Put the binary in a folder listed in your %PATH%. If you don't have an appropriate folder set up, create a new one (ie C:\Lingo) and append it to PATH with a ; in between by going to Control Panel\System and Security\System -> Advanced system settings -> Environment Variables

You must run lingo from Git Bash or similar environment (if not git bash, then at least with git and msys installed). Running in Cmd will not work.

### Linux / Unix

Place the lingo binary on your $PATH.

## Setup

(Note, these steps will be automated in a future release)

1. Create a CodeLingo account: [http://codelingo.io:3030/user/sign_up](http://codelingo.io:3030/user/sign_up)

2. Create a remote repositoy [http://codelingo.io:3030/repo/create](http://codelingo.io:3030/repo/create)

(Note, for now the repository MUST be public)

3. Add a new or existing local repository as a remote:

  ```bash
  touch README.md
  git init
  git add README.md
  git commit -m "first commit"
  git remote add codelingo http://codelingo.io:3030/<your-username>/<your-repo-name>.git
  Git push -u codelingo master
  ```

  It's important that the remote is called “codelingo”, as this is how the lingo client knows which repository to sync with on codelingo.io.

4. Initiate lingo:

```bash
$ lingo init
```

This will write a .lingo file in the current directory.

## Run a Review

The `lingo` tool uses Tenets to review code. Tenets live in .lingo files. The `$ lingo init` command adds a simple Tenet which finds all functions to get you started. To test this Tenet add a file, named “test.php”, with the following source code:

```PHP
<?php
function writeMsg() {
    echo "Hello world!";
}

writeMsg(); // call the function
?>
```

Then run `$ lingo review`. You should see the following output:

```bash
test.php:2

    This is a function, but you probably already knew that.
    

    ...
    
  > function writeMsg() {
        echo "Hello world!";
    }
    
    ...

[o]pen [d]iscard [K]eep: 
```

To open the test.php file at the line of the issue, type `o` and hit return. It will give you an option (which it will remember) to set your editor, defaulting to vi.

## Write a Tenet

Run `$ vi .lingo` and update .lingo to the following:

```yaml
tenets:
- name: first-tenet
  comment: This is a function, name 'writeMsg', but you probably knew that.
  match: 
    <func:
      name: "writeMsg"
```

This will find funcs named "writeMsg". Save and close the file, then run `$ lingo review`. Try adding another func called "readMsg" and run a review. Only the "writeMsg" func should be highlighted. Now, update the Tenet to find all funcs that end in "Msg":

```yaml
  match: 
    <func:
      name: /.*Msg$/
```

The "<" symbol returns the node that you're interested in. The review comment is attached to the returned node. There can only be one returned node per match statement. If a match statement has no "<", then even if true, no issue will be raised.

## CLQL

CLQL is the query language under the `match:` section of a Tenet. It stands for CodeLingo Query Language. The full spec can be found [here](https://docs.google.com/document/d/1NIw1J9u2hiez9ZYZ0S1sV8lJamdE9eyqWa8R9uho0MU/edit), but a practical to get acquainted with the language is to review the [examples](codelingo/lingo/examples).

Release v0.1.0 is a partial implementation of CLQL. String and regex assertions, as demonstrated above, are supported against the following facts in PHP:

```yaml
func:
  name:

call:
  name:

arg:
  name:
  type:

class:
  name:

var:
  name:

literal:
 value:
 type:
```

### Types

"string", "integer", "float", "boolean", "null", "resource", "array", "object", "function", "invalid-type" can be passed to type fact e.g.

```yaml
match:
  literal:
    type: "string"

```

Other than the match statement, written in CLQL, the rest of a .lingo file is written in YAML. As such, you can set .lingo files to YAML syntax in your IDE to get partial highlighting. Vim has full support for the Lingo syntax, including CLQL. To set it up, [see here](scripts/lingo.vim.readme).

## Running Examples

All examples under [examples/php](examples/php) are working. The other examples have varying levels of completeness and serve as an implementation roadmap. To run the examples, copy the directory out of the repository and follow the same steps as in the tutorial above.