# CI/CD Demo

A small Go API I'm using to learn about CI/CD and deployments. This is a personal project, made public to show what I've tried. I'm still figuring things out, so it will probably change.

The Gin API stores a service catalog in memory. Data disappears on restart, and the catalog doesn't manage actual services. Deployment infrastructure lives in the separate [cicd-infra repository](https://github.com/nabilfikrisp/cicd-infra).

## Progress so far

1. Built the API with a replaceable repository, tests, linting, a multi-stage Dockerfile, and a local Compose file.
2. Configured a `main` ruleset that requires pull requests, blocks deletion and force pushes, and has no bypass actors. It requires no approvals because I can't approve my own PRs.
3. Restricted pull request creation to collaborators. The repository is public so people can see the work, but it isn't intended as a shared project.
4. Added a GitHub Actions workflow for linting and race tests on pull requests and pushes to `main`. Those checks have read-only repository permissions and no deployment credentials. Action version tags can move, so their code can change over time.
5. Verified both checks on GitHub and added `lint` and `test` as required checks in the `main` ruleset.
6. Published a commit-SHA-tagged image to GHCR and pulled it successfully. The package is public so others can try the demo without registry credentials. Publishing still requires permission; the planned Arcane dashboard and deployment webhook remain private.

If I make the package private later:

- The publishing job can keep using `GITHUB_TOKEN` with `packages: write`, as long as this repository's Actions workflows retain write access to the package. Check the package's Actions access settings if publishing fails.
- Server (Arcane) will need GHCR registry credentials from an account with access to the package and a token with `read:packages` so it can pull the image. Configure the credential in Arcane and keep the token outside Git. Pulling from other machines will also require authentication. I haven't tested private GHCR pulls through Arcane yet.

## What's next

After checks pass on `main`, the publish job builds one image and pushes it with both the full commit SHA and `main` tags. It then joins Tailscale using GitHub OIDC and sends Arcane a redeploy request. The job succeeds when Arcane returns HTTP 202; Arcane reports deployment failures separately. This automated flow still needs end-to-end verification.

The publish and notification steps share a concurrency lock. A run checks that its commit is still the `main` branch head before pushing the moving tag, so stale queued or rerun workflows fail rather than replace it. GitHub may replace pending runs, and the lock ends when Arcane accepts the request; a later publication may therefore change the image Arcane pulls. Only this workflow's publish job participates in the lock.

Configure repository secrets `TS_OIDC_CLIENT_ID`, `TS_OIDC_AUDIENCE`, and `ARCANE_WEBHOOK_URL`. The Tailscale federated identity must trust the repository's `main` branch, have Auth Keys write access for `tag:runner`, and the tailnet policy must allow that tag to reach Arcane on TCP 443. The webhook URL must use HTTPS on port 443. Only the publish job has package write and OIDC permissions.

To roll back, change the image in the VPS Compose project to `ghcr.io/nabilfikrisp/cicd-demo:<full-commit-sha>` for a known good release, then redeploy it through Arcane. Restore `:main` when ready to resume following new builds; a later automated deployment can otherwise replace the rollback.

Dependency, image, and secret scanning will come later, so images aren't automatically checked for known vulnerabilities yet.

## Run and test

Requires Go 1.27.3 or later and Make. Linting also requires golangci-lint v2.

```bash
make run       # Listens on port 3003; set PORT to change it
make test      # Runs tests with the race detector
make lint
```

Or run with Docker Compose:

```bash
docker compose up --build
```

Compose checks the API's `/health` endpoint and reports the container's health status.

Try the API:

```bash
curl -X POST http://localhost:3003/services \
  -H 'Content-Type: application/json' \
  -d '{"name":"billing-api","description":"Handles invoices"}'
curl http://localhost:3003/services
curl http://localhost:3003/health
curl http://localhost:3003/version
```

Names are required, unique, case-sensitive, and limited to 200 characters. Descriptions are optional and limited to 1,000 characters. Writes don't require authentication.

To build a versioned image:

```bash
make image VERSION=1.0.0 COMMIT="$(git rev-parse HEAD)" IMAGE=cicd-demo:1.0.0
docker run --rm -p 3003:3003 cicd-demo:1.0.0
```

`/version` reports the supplied version and commit, or `dev` and `unknown` by default. A local build can include uncommitted changes that the commit SHA doesn't describe.

The Docker image defaults to Gin's release mode. Set `GIN_MODE=debug` when running the container to enable debug mode; local runs use Gin's default unless `GIN_MODE` is set.
