# tmpl-renderer

## Why

The idea behind this little but very useful project was to somehow have the possibility to include Kubernetes Secrets in source control repositories.

## Installation

`go get github.com/glerchundi/tmpl-renderer/...`

## Usage

Imagine you have the current Secret spec:

```
apiVersion: v1
kind: Secret
metadata:
  name: lb-dhparam
type: Opaque
data:
  dhparam.pem: LS0tLS1CRUdJTiBESCBQQVJBTUVURVJTLS0tLS0KTUlHSEFvR0JBSjVJVFVoMC9OY2JJZXpLcHJpUlMyalJnWDhnZDBqdVNjMFZDai9MOGdCR1hXaGhGQ3Jnb2dTUApST2dGdkxoeUYxU1ZSR0UyL2Y5eEpiTEMwKzczSWdUU0tsdEF5bTBKemh5MkxWQ1FoT2llOU5lWVo3eksrZ3pICi9McGdiS1U4K1VVOHpaQk93aGd0QlpYUHdyZ2duMnBnMG5uMUo0Y2hpaDR2U3EvSTkwanJBZ0VDCi0tLS0tRU5EIERIIFBBUkFNRVRFUlMtLS0tLQo=
```

Obviously having this commited in any repo is a **BIG NO**.

With `tmpl-renderer` you can do the following instead, which of course is safely commitable:

```
apiVersion: v1
kind: Secret
metadata:
  name: lb-dhparam
type: Opaque
data:
  dhparam.pem: {{ getenv "lb-dhparam-dhparam.pem" }}
```

But, now how to emit this template parsed to your Kubernetes cluster, directly? 

1. Create a folder with your secrets inside. I usually create symbolic links so that it's easier to create your custom naming patterns in your templates. Something like this:
  
  ```
  12:33:10 $> ls -l /your-secrets-path
  total 8
  lrwxr-xr-x  1 glerchundi  staff  48 Feb 11 12:33 lb-dhparam-dhparam.pem -> /Volumes/PRIVATE/loadbalancers/staging/dh.pem
  ```
  
2. Execute this awesome snippet which does the following: lists your secrets folder (`/your-secrets-path`), converts every file into an environment variable (`filename=base64(content(filename))`), executes tmpl-renderer (which outputs everything to the terminal) and finally emits the parsed template to Kubernetes (using stdin as input):

```
env $(find /your-secrets-path -maxdepth 1 -type l | xargs -I . sh -c "printf '%s=%s ' \$(basename '.') \$(base64 -i '.')") tmpl-renderer secrets.tmpl.yml 2>&1 | kubectl create -f -
```

**NOTE 1**: Please first test if your template works as expected by removing the `kubectl` call!

**NOTE 2**: `-type l` looks for symbolic links, you can modify it to look for regular files using `-type f`.

## Building

Define building configuration:
```
export VER=$(./git-version)
export VER_TRIM=$(echo -e "${VER}" | sed -e 's/^[[v]]*//')
export APP="tmpl-renderer"
```

Start building:
```
docker build --build-arg version=${VER_TRIM} -f Dockerfile.build -t tmpl-renderer:build-${VER} .
docker run --rm tmpl-renderer:build-${VER} \
   cat /go/src/github.com/glerchundi/tmpl-renderer/bin/tmpl-renderer-${VER_TRIM}-linux-amd64 \
   > bin/tmpl-renderer-${VER_TRIM}-linux-amd64
chmod 0755 bin/tmpl-renderer-${VER_TRIM}-linux-amd64
```
