## mock.json

This file is mocking out the results we will get from the lexicons. Use the lingo binary in demo/lingo. For example, if you call `$ lingo review simple.go` in examples/simple, the following mock data is used:

```json
    "simple": {
        "lexicons": [{
            "name": "codelingo/common",
            "facts": [{
                "name": "func",
                "nodeprops": {
                    "name": "NewReallyLongNamedFunc",
                    "length": 10
                },
                "nodeedges": {
                    "fileedge": {
                        "Filename": "simple.go",
                        "StartLine": 3,
                        "EndLine": 5
                    }
                }
            }]
        }]
    }
```

`simple` is the name of the dir lingo is reviewing in. `lexicons` is a list of mock lexicons. Each lexicon has a list of facts, each fact has a name and returns a node. `nodeprops` and `nodeedges` populates the node returned by the fact. A node can have different edges (e.g. a Lingo AST edge, or a Golang AST edge) - in the above we mock out a fileedge so we can connect the node to a position in a file.