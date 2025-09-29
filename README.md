# first.fm

your last.fm stats within Discord, isn't it great?

the code is work in progress so if u see a war crime just ignore it, thanks. ~elisiei

## installation

### clone repo (via http or ssh)

```sh
$ git clone https://github.com/nxtgo/first.fm
```

### create a .env

```env
DISCORD_TOKEN=token_here
LASTFM_API_KEY=your_api_owo
```

### run using Makefile

```sh
$ make build; make run
```

running without Make **won't work** as Makefile loads
env file. if you want to avoid using Make, pass the
env variables on command invokation or use a [go env loader](https://github.com/nxtgo/env).

# license

all original content in this project is dedicated to the public domain under the
[CC0 1.0 universal](https://creativecommons.org/publicdomain/zero/1.0/).
this grants you the freedom to use, modify, and distribute the content
without any restrictions or attribution requirements.
