# CI/CD Demo

A small Go API for learning CI/CD and deployments. The Gin API stores a service catalog in memory; data disappears on restart, and the catalog doesn't manage actual services. Deployment infrastructure lives in the separate [cicd-infra repository](https://github.com/nabilfikrisp/cicd-infra).

## CI/CD

Pull requests and pushes to `main` run lint and race tests with read-only permissions and no deployment credentials. The `main` ruleset requires both checks and pull requests, blocks deletion and force pushes, and has no bypass actors or required approvals. Pull request creation is restricted to collaborators. Actions use version tags, which can move.

After checks pass on `main`, the publish job builds one image, pushes it to GHCR with both the full commit SHA and `main` tags, joins Tailscale using GitHub OIDC, and sends Arcane a redeploy request. The job succeeds when Arcane returns HTTP 202; Arcane reports deployment failures separately. A redeploy has been confirmed successful, and the public `/health` and `/version` endpoints work through Cloudflare Tunnel. HTTP 202 alone does not confirm deployment completed.

The publish and notification steps share a concurrency lock. A run checks that its commit is still the `main` branch head before pushing the moving tag, so stale queued or rerun workflows fail rather than replace it. GitHub may replace pending runs, and the lock ends when Arcane accepts the request; a later publication may therefore change the image Arcane pulls. Only this workflow's publish job participates in the lock.

The working setup uses repository secrets `TS_OIDC_CLIENT_ID`, `TS_OIDC_AUDIENCE`, and `ARCANE_WEBHOOK_URL`. The Tailscale identity trusts this repository's `main` branch and has Auth Keys write access for `tag:runner`; the tailnet policy allows that tag to reach Arcane on TCP 443. The dashboard and webhook are private. Only the publish job has package write and OIDC permissions.

The GHCR package is public. If made private:

- The publish job can use `GITHUB_TOKEN` with `packages: write` while this repository's Actions retain package write access. Check the package's Actions access settings if publishing fails.
- Arcane needs credentials from an account with package access and a `read:packages` token. Configure them in Arcane and keep them outside Git. Other machines also need authentication. Private pulls through Arcane have not been tested.

### Tailscale OIDC subject

The first token exchange failed with HTTP 403 because the Tailscale credential expected a name-only GitHub subject, while GitHub sent the immutable subject containing the owner and repository IDs. For this repository's push to `main`, the matching subject is:

```text
repo:nabilfikrisp@79518623/cicd-demo@1471480995:ref:refs/heads/main
```

If token exchange fails, open the OIDC credential under Tailscale **Trust credentials**, select **View or Edit**, and check the latest error and the subject received from GitHub. Match the credential's subject to that value. GitHub's subject format varies with repository settings and the workflow context; a GitHub Actions environment, for example, changes the subject. See [GitHub's OIDC reference](https://docs.github.com/en/actions/reference/security/oidc#immutable-subject-claims).

### Arcane webhook URL

Arcane's webhook dialog may show a URL using its HTTP IP address and port. For `ARCANE_WEBHOOK_URL`, replace that origin with the private HTTPS hostname served by Tailscale, keeping Arcane's complete webhook path and token unchanged. For example:

```text
http://<arcane-ip>:<port>/<webhook-path-and-token>
https://<machine>.<tailnet>.ts.net/<webhook-path-and-token>
```

Tailscale Serve must route the HTTPS hostname on port 443 to Arcane's HTTP service. Check the existing configuration with `tailscale serve status`; see the [Tailscale Serve guide](https://tailscale.com/docs/solutions/host-private-server-docker) when setting it up. Keep the actual webhook URL and token in the GitHub secret, not in the repository or logs.

To roll back, change the image in the VPS Compose project to `ghcr.io/nabilfikrisp/cicd-demo:<full-commit-sha>` for a known good release, then redeploy it through Arcane. Restore `:main` when ready to resume following new builds; a later automated deployment can otherwise replace the rollback. This rollback procedure has not yet been tested.

## What's next

Test rollback and deliberately failed publication and notification in a live workflow. The failure paths have been checked locally with mocked commands. Dependency, image, and secret scanning will come later, so images aren't automatically checked for known vulnerabilities yet.

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
