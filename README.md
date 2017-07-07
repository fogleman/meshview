# meshview

Performant 3D mesh viewer written in Go.

### Prerequisites

First, install Go, set your `GOPATH`, and make sure `$GOPATH/bin` is on your `PATH`.

```
brew install go
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"
```

You may need to [install prerequisites](https://github.com/go-gl/glfw#installation) for the glfw library.

### Installation

```
go get -u github.com/fogleman/meshview/cmd/meshview
```

### Usage

```bash
$ meshview model.stl
```

![Screenshot](http://i.imgur.com/6RKNQuf.png)
