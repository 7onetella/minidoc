# minidoc

Manage small pieces of information using minidoc. Currently, you can manage URLs, TODOs, and Notes. 

The problem minidoc is trying to solve is storing and retrieving small pieces of information that's everywhere on computers. full text search is enabled so it makes it easy to find the stored information later on.

Another useful feature of minidoc is batch tagging. You can search for information, filter and then apply tags to multiple minidocs. Selected minidocs can be used to generate markdowns or pdfs if pandoc is installed. 

## Installation using homebrew
```console
$ brew tap 7onetella/minidoc
$ brew install minidoc
$ minidoc
```

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

Running it then should be as simple as:

```console
$ cd minidoc
$ go build && ./minidoc
```

### Testing
TODO: boltdb is used so testing seems to block

## Demo - Asciicast
[![asciicast](https://asciinema.org/a/MoSChtTE6KuLhzg4w0TJl8Puv.svg)](https://asciinema.org/a/MoSChtTE6KuLhzg4w0TJl8Puv)

## Design decision

I am mostly writing this app for myself to solve my own problmes. It's not meant to be an online app. It's designed for single user who like text user interface. If you really need to share your miniocs with someone else you can export and have that person import. If someone wanted a full gui app that does what minidoc does, then there are plenty of apps out there that are cloud native. 


