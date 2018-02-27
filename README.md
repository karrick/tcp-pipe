# tcp-pipe

tcp-pipe: a light-weight UNIX pipe via TCP socket

## Description

The tcp-pipe program allows you to stitch together STDOUT on one
machine to STDIN on another machine through a TCP connection.

Why not the venerable `netcat`, a.k.a., `nc`? Indeed, `nc` is a
versatile and capable tool. However when using in a distributed
environment, some versions of `nc` failed to terminate once their
STDIN was closed, causing the network connection on both ends to
hang. If you experience the same problems with your `nc` pipe,
consider giving `tcp-pipe` a try.

## Usage Example

### Sending a directory hierarcy to host.example.com

Why not use the amazing `rsync` program? You should! But really, if
you're running on hosts where you cannot easily install dynamically
linkedin binaries, or when you are running without the need for
encryption, this can be a handy alternative.

1. First, run this on the destination host:

    tcp-pipe receive :6969 | tar xf -

1. Second, run this on the source host:

    tar -c --format pax -f - someDir | tcp-pipe send host.example.com:6969

Among other benefits, using the `pax` format allows for better
handling of pathnames.
