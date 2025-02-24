# storage-manager

Per-worker-node storage manager, responsible for retrieving aimages, models and components, storing them locally and making them available to inference engines.

## Responsibilities

- Downloading and storing images, models and components to a configurable cache directory
- Managing everything optimally in the cache using [OCI layout format](https://github.com/opencontainers/image-spec/blob/main/image-layout.md), which provides content-addressable storage (CAS) and deduplication
- Exposing an API to pass it commands:
  - Download a specific aimage, model or component from a provided URL
  - Check if a specific aimage, model or component is available and complete
  - Remove a specific aimage, model or component from the cache
- Enable configuration of options via CLI flags and environment variables, with reasonable defaults

## Configuration Options

| Option | Flag | Env Var | Description | Default |
| ------ | ---- | ------- | ----------- | ------- |
| Cache Directory | `--cache-dir` | `CACHE_DIR` | Directory where images, models and components are stored | `/var/lib/nekko/cache` |
| Address | `--address` | `NEKKO_ADDRESS` | Address and port or Unix-domain socket where the API listens | `localhost:8050` |
| Log Level | `--verbose` | `VERBOSE` | Log level for the application | `0` |

## API

The storage manager exposes an API with the following endpoints:

- `GET /content/<URL>`: Check if URL is available in cache.
- `POST /content/`: Download content from the provided URL and store it in the cache.
- `DELETE /content/<URL>`: Removes content from the cache.

### GET /content/<URL>

Check if a specific URL is available in the cache. Returns `200` if the provided content URL is available in the cache. Returns `404` if not available, `200` if available and complete, and `206` if available but incomplete. URL is base64-encoded.

### POST /content/

Downloads content from the provided URL and stores it in the cache. Body
contains json with the URL to the content. Returns `201` if successful.

Body is as follows:

```json
{
  "url": "<URL>"
}
```

URL is **not** base64-encoded.

The URL format determines which downloader is used.

You can provide credentials via the field `"credentials"` and an optional `"credentialsType"` field. E.g.:

```json
{
  "url": "<URL>",
  "credentials": "<TOKEN>"
}
```

or

```json
{
  "url": "<URL>",
  "credentials": "<TOKEN>",
  "credentialsType": "Bearer"
}
```

The interpretation of the token is up to the individual downloader.

### DELETE /content/<URL>

Removes the aimage from the cache. URL is base64-encoded. Returns `204` if successful.

## Downloaders

The following downloaders and request formats are supported.

### OCI

* URL format: `oci://<registry>/<repository>/<image>:<tag>` or `oci://<registry>/<repository>/<image>@<digest>`
* Credentials: token
* Credentials Type: Only `Bearer` supported, defaults to `Bearer`

### HuggingFace

* URL format: `huggingface://<registry>/<model>/<file>` or `hf://<registry>/<model>/<file>`; if no `<registry>` is supplied, defaults to `huggingface.co`, e.g. `hf:///unsloth/SmolLM2-135M-Instruct-GGUF/SmolLM2-135M-Instruct-Q2_K.gguf` (note three `/` following `hf`)
* Credentials: token
* Credentials Type: Only `Bearer` supported, defaults to `Bearer`

### http

Supports both http and https

* URL format: `http://<host>/<path>` or `https://<host>/<path>`
* Credentials: token or username-password, `:`-separated and base64-encoded
* Credentials Type: `Bearer` or `Basic`, defaults to `Bearer`
