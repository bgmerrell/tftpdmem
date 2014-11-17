tftpdmem
========

tftpdmem is a TFTP server that stores its files as buffers in memory (though  Nothing actually prevents the OS from swapping the memory buffers to disk).

Getting and Running
========

Assuming you have a Go environment set up, simply run the following:

```go get github.com/bgmerrell/tftpdmem```

And then run (assuming you want to run the server on port 6969):

```$GOPATH/bin/tftpdmem --port 6969```

If you don't have a Go environment setup, please follow the instructions over at https://golang.org/doc/code.html first.
