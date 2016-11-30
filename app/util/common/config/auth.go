package config

var AuthTmpl = `
all:
  gitserver:
    credentials_filename: git-credentials
    user:
      password: ""
      username: ""
dev:
  gitserver:
    credentials_filename: git-credentials-dev
test:
  gitserver:
    credentials_filename: git-credentials-test`[1:]
