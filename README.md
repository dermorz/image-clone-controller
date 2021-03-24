# image-clone-controller

A Kubernetes controller which watches the applications and "caches" the images
by re-uploading to our own registry repository and reconfiguring the
applications to use these copies.

## Demo

The controller in action: https://asciinema.org/a/O7Qk6BiZaoAiVMUSaXkOBkOdM

You can actually spot it doing its thing on itself directly after deploying.
But after that the functionality is demonstrated on an example nginx
deployment.

## Prerequisites

* Kubebuilder (Project was scaffolded with version `3.0.0-beta.0`. Earlier version might not work.)
* docker (>=`19.0`)

## Setting up the Backup Registry

Either create a file `./config/manager/secrets/.dockerconfigjson` with the content:

```json
{
  "auths": {
    "<registry>": {
      "username": "<username>",
      "password": "<password>",
      "email": "<email>",
      "auth": "<base64 encoded '<username>:<password>'>"
    }
  }
}
```

or just run `tools/generate-pull-secret` and follow the prompt.

## How to deploy

```shell
$ export IMG=<registry>/<image>:<tag>
$ make docker-build
$ make docker-push
$ make deploy
```

## Known bugs

* Some source images when retagged trigger a `401 Unauthorized` when pushed to
  the "backup registry".
* No testing

## Disclaimer

This is my first Kubernetes controller, so this probably is far from best
practices. There are always things to improve, but learning to write a
controller by doing it while working and having 2 kids at home during corona
are the reasons why I kept the scope pretty narrow and went for the naive
approach.

Nevertheless it was a great learning experience. `kubebuilder` took care of a
lot of scaffolding and also helped leading the way.

I still have no clue how to properly test a controller. :(
