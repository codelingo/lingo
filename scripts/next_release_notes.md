This release adds a review subcommand, pull-request, for reviewing remote pull requests on github. It can be tested with the following:

```bash
$ lingo review pull-request https://github.com/waigani/codelingo_demo/pull/6
```

 A --lingo-file flag was added which allows .lingo files to be passed in on the command line instead of reading .lingo files from the target repository.

 The release script was updated to generated change logs.