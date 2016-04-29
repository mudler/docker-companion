# docker-companion [![Build Status](https://travis-ci.org/mudler/docker-companion.svg?branch=master)](https://travis-ci.org/mudler/docker-companion)

docker-companion is a candy mix of tools for docker written in Golang and directly using Docker API calls. As for now it allows to squash and unpack an image.

## Reinventing the wheel?

Problem arises with current tools to squash/unpack images since mostly of them are scripted. I personally needed a static implementation with no-deps hell that i could use in my CI pipeline easily (and also to get the job done).

## Squash an image

The resulting image will loose metadata, but it is handy to reduce image size:

    docker-companion squash my-awesome-image my-awesome-image-squashed

You can also make it pull before squashing it:

    docker-companion --pull squash my-awesome-image my-awesome-image-squashed:mytag

## Unpack an image

    docker-companion unpack my-awesome-image /my/path

The path must be absolute, and you must run it with root permission to avoid keeping permissions to the files.

You can squash the image right before unpacking it too:

    docker-companion --pull unpack --squash my-awesome-image /my/path

It can be handy sometimes to squash the image before unpacking it (very few cases where the latter fails)
