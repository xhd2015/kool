# __NAME__

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/__OWNER__/__REPO__/main/install.sh | bash
```

To install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/__OWNER__/__REPO__/main/install.sh | INSTALL_TAG=v1.0.0 bash
```

## Release

To release a new version to GitHub Releases:

1. Commit all changes and tag the commit:
   ```sh
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. Create `.upload-credentials.json` with your GitHub token:
   ```json
   {"token": "ghp_...", "owner": "__OWNER__", "repo": "__REPO__"}
   ```

3. Run the release script:
   ```sh
   go run ./script/release
   ```

Or do a dry run first:
```sh
go run ./script/release --dry-run
```
