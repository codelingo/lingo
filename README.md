
<h3 align="center"> codelingo </h3>

<p align="center">
  <b> Automate Your Reviews on GitHub Pull Requests! </b>
</p>

<p align="center">
  <a href="https://github.com/apps/codelingo" target="_blank">
    <img width="295" height="38" src="https://raw.githubusercontent.com/codelingo/codelingo/master/public/img/install.png" />
  </a>
</p>

# Lingo

Lingo is a CLI tool to run [CodeLingo](https://www.codelingo.io) with any git repository on your local machine.

### Quick Start

In this quick start we'll review a Golang git repository on your local machine for common Go issues.

After installing the [lingo tool](https://github.com/codelingo/lingo/releases/latest), set it up with the following commands:

```bash
# Run this command from anywhere. Follow the prompts to set up Codelingo on your machine.
$ lingo config setup

# Run this command inside a git repository to add a default codelingo.yaml file in the current directory.
$ lingo init
```

Replace the default content of the codelingo.yaml file we generated above with:

```yaml
  tenets:
    - import: codelingo/go
```

Run the Review Flow to check your source code against the go Tenet bundle we imported above:

```bash
# Run this command from the same directory as the codelingo.yaml file or any of its sub directories.
$ lingo run review
```

<!-- TODO add screenshot of lingo review -->

## Slow Start

Follow the [step by step guide](https://www.codelingo.io/docs/getting-started) to using lingo.

See and <a class="github-button" href="https://github.com/codelingo/codelingo" data-icon="octicon-star" aria-label="Star codelingo/codelingo on GitHub">star</a> the main repository at [github.com/codelingo/codelingo](https://github.com/codelingo/codelingo).