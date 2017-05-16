# Micro-services app with nanobox

## Overview

This repo is meant to serve as one example of how multiple codebases can be combined into a single nanobox application. Each 'service' has it's own boxfile and can be developed individually with nanobox. This example uses golang applications and the process will likely vary for any other language. The directory structure is as follows

```sh
/admin  # admin service
  /app  # admin content

/site   # site service
  /app  # site content

/store  # store service
  /app  # store content

/worker # worker service
```

For this example, each directory contains a simple golang project and nanobox `boxfile.yml` allowing the user to build the application binary with `nanobox run go build`.

The outermost boxfile.yml tells nanobox to use the `none` engine and includes the necessary commands to ensure the complete application ends up on the provider directory for `nanobox deploy`. It also includes special routing rules to direct traffic to the correct component.
```yaml
# Don't have nanobox automatically configure the environment
run.config:
  engine: none

# Copy the code into the proper location for `nanobox deploy`
deploy.config:
  extra_steps:
    - cp -r /app/* $APP_DIR/

data.cache:
  image: nanobox/redis

web.admin:
  start: "/app/admin/admin /app/admin/app"
  routes: # Route `admin.myurl.com` and `myurl.com/admin` to this component
    - '/admin'
    - 'admin:/'

web.site:
  start: "/app/site/site /app/site/app"
  routes:
    - '/'

web.store:
  start: "/app/store/store /app/store/app"
  routes: # Route `store.myurl.com` and `myurl.com/store` to this component
    - '/store'
    - 'store:/'

worker.worker:
  start: "/app/worker/worker"
```

--------

#### Run composite app locally
```sh
nanobox deploy dry-run

# add the dns in another terminal
nanobox dns add dry-run multi-app.dev
nanobox dns add dry-run admin.multi-app.dev
nanobox dns add dry-run store.multi-app.dev
```

#### Develop and build sub applications
```sh
cd admin
nanobox run

# develop and test the application as needed

go build
exit
```

--------

### DISCLAIMER
I make no guarantee as to the condition of this example application or guide.
