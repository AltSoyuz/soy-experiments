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

