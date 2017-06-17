# prs

prs is a small cli utility for listing open pull requests from github that you 
are involved in.

## usage

prs knows about two environment variables `PRS_GITHUB_ACCESS_TOKEN` and
`PRS_USERNAME`. If invoked without parameters the username from the
environment variable is used. A parameter with a username can be supplied to
the invocation to override the environment variable.

```
$ prs [PRS_USERNAME]
```
