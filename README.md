# meshview

Performant 3D mesh viewer written in Go.

### Installation

First, install Go, set your `GOPATH`, and make sure `$GOPATH/bin` is on your `PATH`.

```
brew install go
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"
```

Next, fetch and build the `meshview` binary.

```
go get -u github.com/fogleman/meshview/cmd/meshview
```

### Usage

```bash
$ meshview model.stl
```

![Screenshot](http://i.imgur.com/6RKNQuf.png)
