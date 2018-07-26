# CodeLingo - Code Quality That Scales

## Overview

CodeLingo helps software development teams produce better software, together. It's a platform that supports quering software as data, with CLQL (CodeLingo Query Language) and automating development workflows, called Flows.

Our flagship Flow is the Review Flow, which checks pull requests to a repository conform to the project specific patterns of that repository. Other Tenet bundles (including for other languages) from the community can be found under the [tenets directory](https://github.com/codelingo/hub/tree/master/tenets) in the [https://github.com/codelingo/hub](https://github.com/codelingo/hub) repository.

## Quick Start: GitHub Review Flow

1. Install the [CodeLingo GitHub App](link)

2. Write the following .lingo.yaml to the root of your repo:


      ```yaml
      tenets:
         - import: codelingo/go
      ```

3. Done! Every pull request will now be checked against the go Tenet bundle we imported above. 

## Quick Start: Local Review Flow

The Review Flow can also be run against repositories on your local machine.

1. Install the [lingo CLI](github.com/codelingo/lingo)

2. Run the following commands:

```bash
# Run this command from anywhere. Follow the prompts to set up CodeLingo on your machine.
$ lingo config setup

# Run this command inside a git repository to add a default .lingo.yaml file in the current directory.
$ lingo init
```

3. Write the following .lingo.yaml to the root of your repo:


```yaml
  tenets:
    - import: codelingo/go
```

# Add a Tenet to the .lingo.yaml file (see https://codelingo.io/docs/concepts/tenets/#writing-custom-tenets for more info). This will be used by the following command to run a review.
$ lingo run review

```

## Slow Start

See [https://codelingo.io/docs/tutorials/getting-started](https://codelingo.io/docs/getting-started) for more detailed instructions.

## Community

Find us on **[Slack](https://join.slack.com/t/codelingo/shared_invite/enQtMzY4MzA5ODYwOTYzLWVhMjI1ODU1YmM3ODAxYWUxNWU5ZTI0NWI0MGVkMmUwZDZhNWYxNGRiNWY4ZDY0NzRkMjU5YTRiYWY2N2FlMmU)** and [codelingo.io/discuss](codelingo.io/discuss)