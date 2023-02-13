# Contributing

Thank you for your interest in the Instaclustr Operator!
We are interested in all contributions, big or small, in code and/or documentation.

### Requirements

- Golang
- Docker
- Kubectl
- Kubebuilder
- Access to a Kubernetes cluster

### The working process

Leave a comment on an existing issue if you want to work on it.
If you are going to fix a bug or add a feature, please create an issue for it.
This way, we will avoid unnecessary efforts in case something can't be worked on right now
or someone is already working on it.

### Code style and conventions

We value consistency in the source code.
If you see some code that doesn't follow some best practice but is consistent,
please keep it that way; but please also tell us about it, so we can improve it.

### Submitting code changes

Before submitting a pull request or re-request review, please make sure that:
1. You have added and/or updated comments for all new changes (functions, types, vars, etc).
2. If you have changed CRD, please run `make manifest`, `make generate`.
3. You run `make test` without errors.
4. You replied to reviewers comments.
5. You linked an issue in a PR description.
6. You added proper labels on your PR. 
