package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

  "github.com/midbel/hexdump"
)

func main() {
	var (
		block   = flag.Int("b", 0, "read block of bytes")
		bits    = flag.Bool("x", false, "print byte in bits instead of hex")
		size    = flag.Int64("n", 0, "read number of bytes")
		skip    = flag.Int64("s", 0, "skip number of bytes")
		width   = flag.Int("w", 0, "columns width")
		cols    = flag.Int("c", 0, "columns")
		padding = flag.String("p", "   ", "padding")
		delim   = flag.String("d", "|", "delimiter")
		verbose = flag.Bool("v", false, "verbose")
	)
	flag.Parse()

	options := []hexdump.Option{
		hexdump.WithPadding(*padding),
		hexdump.WithDelim(*delim),
		hexdump.WithColumns(*cols),
		hexdump.WithWidth(*width),
		hexdump.WithVerbose(*verbose),
		hexdump.WithBits(*bits),
	}
	d := hexdump.New(options...)

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Close()

	if *block <= 0 {
		*block = d.BlockSize()
	}

	var (
		buf = make([]byte, *block)
		rs  = bufio.NewReader(LimitReader(r, *size, *skip))
	)
	for {
		n, err := rs.Read(buf)
		if err != nil {
			break
		}
		str := d.Dump(buf[:n])
		fmt.Println(str)
	}
}

func LimitReader(r io.ReadSeeker, size, skip int64) io.Reader {
	if skip > 0 {
		r.Seek(skip, io.SeekStart)
	} else if skip < 0 {
		r.Seek(skip, io.SeekEnd)
	}
	if size <= 0 {
		return r
	}
	return &io.LimitedReader{R: r, N: size}
}
