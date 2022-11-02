# fc-provision
A container provisioning engine

## How it Works
- Your container has env variables exposed called `FC_BUILDER_CACHE_USERNAME` and `FC_BUILDER_CACHE_PASSWORD`
- The builder will expose an API to interact with
- Make requests to each endpoint with the format http://username:password@url-to-builder/endpoint

## Endpoints
- Build: the Build endpoint takes a `repo_name`, `owner_name`, `cache_url` (optional) and `cookie` field, then attempts to build a container based on the provided GitHub repository. The user must be authenticated.

## Supported Languages
- Go
- Ruby
- Node.JS
- Rust
- Python

## Requirements
- A Kubernetes cluster
- A running fc-session-cache container
- A Docker registry in the cluster with HTTPS enabled

## Docker
The builder is provided as a docker container:
- https://hub.docker.com/r/sthanguy/fc-provision
