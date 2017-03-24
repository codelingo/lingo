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
onprem:
  gitserver:
    credentials_filename: git-credentials-onprem
test:
  gitserver:
    credentials_filename: git-credentials-test`[1:]
