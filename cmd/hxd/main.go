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
		group   = flag.Int("g", 0, "group bytes")
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
		hexdump.WithGroup(*group),
	}
	d := hexdump.New(options...)

	if *block <= 0 {
		*block = d.BlockSize()
	}

	buffer := make([]byte, *block)
	for i, a := range flag.Args() {
		if i > 0 {
			fmt.Println()
		}
		if flag.NArg() > 0 {
			fmt.Println(a)
		}
		if err := DumpFile(d, a, buffer, *size, *skip); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
}

func DumpFile(d *hexdump.Dumper, file string, buf []byte, size, skip int64) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	d.Reset()

	rs := bufio.NewReader(LimitReader(r, size, skip))
	for {
		n, err := rs.Read(buf)
		if err != nil {
			break
		}
		str := d.Dump(buf[:n])
		fmt.Println(str)
	}
	return nil
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
