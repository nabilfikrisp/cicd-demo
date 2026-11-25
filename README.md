# CI/CD Demo

A small Go API I'm using to figure out CI/CD, container builds, and deployments. It's a work in progress, and things will probably change as I learn.

This is a personal learning project, not really something I'm building for other people to use or contribute to. I made it public mostly so I can show what I've tried and what I've learned along the way. Some choices might be awkward or change as I understand things better.

The app is a simple service catalog built with Gin. It stores data in memory, so everything disappears when the app restarts. The catalog is just data; it doesn't deploy or manage services.

This repository is the application used by a separate self-hosted infrastructure project. The plan is to build and publish its image to GHCR, then deploy that image through the infrastructure. The pipeline is still a work in progress.

## CI/CD plans

The CI workflow runs linting and race tests on pull requests and pushes to `main`. It still needs to be verified on GitHub. I'm planning to build and publish an image to GHCR only from trusted pushes to `main` if those checks pass. The image should include the Git commit in `/version` so I can check what is running.

This is a personal learning project, and I intend to be the only person who merges changes into `main`. GitHub's pull request permission is intended to be set to **Creation allowed by collaborators only**, even though the repository is public. I'll keep collaborator access limited to people I trust. Pull request workflows will only lint and test; they won't publish images or have deployment credentials.

The `main` ruleset is active: it requires a pull request, blocks branch deletion and force pushes, and has no bypass actors. It requires zero approving reviews because I can't approve my own pull requests. Required status checks haven't been configured yet; I'll add lint and test checks after the workflow has run on a pull request.

Later, the deployment workflow will temporarily join Tailscale and call an authenticated Arcane webhook. Arcane will pull the published image, so the server won't need an application source checkout or SSH access from the runner. This deployment flow hasn't been verified yet; the infrastructure lives in a separate repository.

Since this is a public repository, I want to have these protections in place before enabling publishing and deployment:

- Protect `main` and changes to workflow files, restrict who can push or merge where the repository's GitHub plan supports it, and require the pull request checks before merging.
- Run pull request checks with read-only permissions and no publishing or deployment credentials. Don't run untrusted pull request code in a privileged workflow.
- Give GitHub Actions the minimum token permissions needed, including package write access only for the publishing job. Actions use readable version tags, so their referenced code can change when those tags move.
- Keep credentials out of the repository and logs. Use short-lived, narrowly scoped OIDC credentials for an ephemeral Tailscale runner and an authenticated, project-scoped webhook if Arcane supports it.
- Publish traceable image tags or digests and plan for rollback and deployment concurrency. Keep Arcane private and expose only the application through Cloudflare Tunnel.

Dependency and image scanning, secret scanning, and possibly image signing are things I want to try after the basic deployment works. Until then, images aren't automatically checked for known vulnerabilities.

## Run locally

You'll need Go 1.27.3 or later and Make:

```bash
make run
```

The server listens on port `3003` by default. Set `PORT` to change it:

```bash
PORT=8080 make run
```

## Try the API

Create a service:

```bash
curl -X POST http://localhost:3003/services \
  -H 'Content-Type: application/json' \
  -d '{"name":"billing-api","description":"Handles invoices"}'
```

List services and check the app:

```bash
curl http://localhost:3003/services
curl http://localhost:3003/health
curl http://localhost:3003/version
```

Names are required, case-sensitive, unique, and limited to 200 characters. Descriptions are optional and limited to 1,000 characters. Writes don't require authentication. The version endpoint returns `dev` and `unknown` unless values are provided at build time.

## Run with Docker

For a local build and run with Docker Compose:

```bash
docker compose up --build
```

Or build an image with version information and run it directly:

```bash
make image VERSION=1.0.0 COMMIT="$(git rev-parse HEAD)" IMAGE=cicd-demo:1.0.0
docker run --rm -p 3003:3003 cicd-demo:1.0.0
```

`COMMIT` should identify the source used for the build. A local build can also include uncommitted changes that aren't represented by that commit.

## Checks

```bash
make test  # Runs Go tests with the race detector
make lint  # Requires golangci-lint v2
```

`make fmt` formats the Go code, and `make build` creates a local binary in `bin/`.

## Structure

The service catalog lives in `internal/modules/service`, with separate domain, contract, handler, use case, and repository packages. The in-memory repository implements the repository contract, so another storage implementation can be wired in later. `cmd/api` wires the application together, and `internal/httpserver` provides shared HTTP routing and operational endpoints.
