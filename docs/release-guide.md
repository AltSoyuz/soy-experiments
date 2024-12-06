## Pre-reqs 

### For MacOS users

Make sure you have GNU version of utilities `zip`, `tar`, `sha256sum`. To install them run the following commands:
```sh
brew install coreutils
brew install gnu-tar
export PATH="/usr/local/opt/coreutils/libexec/gnubin:$PATH"
```

Docker may need additional configuration changes:
```sh 
docker buildx create --use --name=qemu
docker buildx inspect --bootstrap  
```

By default, docker on MacOS has limited amount of resources (CPU, mem) to use. 
Bumping the limits may significantly improve build speed.

## Release version and Docker images

1. Make sure all the changes are documented in [CHANGELOG.md](https://github.com/AltSoyuz/soy-experiments/blob/main/docs/CHANGELOG.md).
   Ideally, every change must be documented in the commit with the change. Alternatively, the change must be documented immediately
   after the commit, which adds the change.

2. Make sure you get all changes fetched `git fetch --all`.

3. Create the following release tags:
   * `git tag v1.xx.y` in `main` branch

4. Run `TAG=v1.xx.y make publish-release`. This command performs the following tasks:
   a) Build and package binaries in `*.tar.gz` release archives with the corresponding `_checksums.txt` files inside `bin` directory.
      This step can be run manually with the command `make release` from the needed git tag.
   b) Build and publish [multi-platform Docker images](https://docs.docker.com/build/buildx/multiplatform-images/)
      for the given `TAG`. 
      The multi-platform Docker image is built for the following platforms:
      * linux/amd64
      * linux/arm64
      This step can be run manually with the command `make publish` from the needed git tag.

5. TODO: Verify that created images are stable and don't introduce regressions on `test environment`. 

6. Push the tags `v1.xx.y` created at previous steps to public GitHub repository at https://github.com/AltSoyuz/soy-experiments.

7. Run `TAG=v1.xx.y make github-create-release github-upload-assets`. This command performs the following tasks:
   a) Create draft GitHub release with the name `TAG`. This step can be run manually
      with the command `TAG=v1.xx.y make github-create-release`.
      The release id is stored at `/tmp/se-github-release` file.
   b) Upload all the binaries and checksums already created. 
      This step can be run manually with the command `make github-upload-assets`.
      It is expected that the needed release id is stored at `/tmp/se-github-release` file,
      which must be created at the step `a`.
      If the upload process is interrupted by any reason, then the following recovery steps must be performed:
      - To delete the created draft release by running the command `make github-delete-release`.
        This command expects that the id of the release to delete is located at `/tmp/se-github-release`
        file created at the step `a`.
      - To run the command `TAG=v1.xx.y make github-create-release github-upload-assets`, so new release is created
        and all the needed assets are re-uploaded to it.

8. Update the release description with the content of [CHANGELOG](https://github.com/AltSoyuz/soy-experiments/blob/main/docs/CHANGELOG.md) for this release.
9. Publish release by pressing "Publish release" green button in GitHub's UI.
