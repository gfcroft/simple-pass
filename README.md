# simple-pass

simple-pass is a simple cli password manager

create a persistent storage medium ('PassDB')
```bash
    simple-pass create-pass-db foobar --password "something at least 5 chars"
```

and then store the usual details (in an encrypted store):
```bash
    simple-pass add eg --username "me" --password "whatever"
``````
## Installation

### using homebrew
```bash
    brew tap gfcroft/homebrew-taps
    brew install simple-pass

```
### other
You can download the binary directly from the releases attached to this repository

or build from source using the makefile
```bash
    make build
```
## Current Status

simple-pass is currently in beta and its api/commands are subject to change 
