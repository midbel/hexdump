package hexdump

import (
	"hash/adler32"
	"strings"
)

const (
	DefaultColumnCount = 2
	DefaultColumnWidth = 8
	DefaultColumnGroup = 1
	DefaultPadding     = "   "
	DefaultDelimiter   = "|"
)

type Option func(*Dumper)

func WithPadding(pad string) Option {
	return func(d *Dumper) {
		d.padding = pad
	}
}

func WithDelim(delim string) Option {
	return func(d *Dumper) {
		d.delim = delim
	}
}

func WithWidth(width int) Option {
	return func(d *Dumper) {
		if width <= 0 {
			return
		}
		d.width = width
	}
}

func WithColumns(cols int) Option {
	return func(d *Dumper) {
		if cols <= 0 {
			return
		}
		d.cols = cols
	}
}

func WithGroup(groups int) Option {
	return func(d *Dumper) {
		if groups <= 0 {
			return
		}
		d.groups = groups
	}
}

func WithBits(bits bool) Option {
	return func(d *Dumper) {
		d.bits = bits
	}
}

func WithVerbose(verbose bool) Option {
	return func(d *Dumper) {
		d.verbose = verbose
	}
}

func Dump(buf []byte) string {
	d := New(WithVerbose(true), WithColumns(5))
	return d.Dump(buf)
}

func Dump2(buf []byte) string {
	d := New(WithVerbose(true), WithColumns(5), WithGroup(2))
	return d.Dump(buf)
}

func Dump4(buf []byte) string {
	d := New(WithVerbose(true), WithColumns(5), WithGroup(4))
	return d.Dump(buf)
}

type Dumper struct {
	width   int
	cols    int
	groups  int
	padding string
	delim   string
	bits    bool
	verbose bool

	size    int
	digest  uint32
	written uint64
	index   int
	buffer  []byte
}

func New(options ...Option) *Dumper {
	d := Dumper{
		width:   DefaultColumnWidth,
		cols:    DefaultColumnCount,
		groups:  DefaultColumnGroup,
		padding: DefaultPadding,
		delim:   DefaultDelimiter,
	}
	for _, opt := range options {
		opt(&d)
	}
	coeff := 2
	if d.bits {
		coeff = 8
	}
	if d.groups > d.width {
		d.groups = d.width
	}
	between := d.width / d.groups
	if mod := d.width % d.groups; mod == 0 {
		between--
	}
	d.size = (d.cols * d.width * coeff) + (d.cols * between) + ((d.cols - 1) * len(d.padding))
	const (
		offlen   = 8
		spacelen = 4
	)
	buflen := offlen + spacelen + d.size + (d.cols * d.width) + (d.cols - 1) + (2 * len(d.delim))
	d.buffer = make([]byte, buflen)
	for i := range d.buffer {
		d.buffer[i] = ' '
	}

	pos := offlen + 1
	copy(d.buffer[pos:], []byte(d.delim))
	pos += d.size + len(d.delim) + 2
	copy(d.buffer[pos:], []byte(d.delim))
	return &d
}

func (d *Dumper) Dump(input []byte) string {
	if !d.verbose {
		sum := adler32.Checksum(input)
		if d.written > 0 && sum == d.digest {
			return "*"
		}
		d.digest = sum
	}
	var (
		str   strings.Builder
		width = d.BlockSize()
	)
	for i := 0; i < len(input); i += width {
		j := i + width
		if j >= len(input) {
			j = len(input)
		}
		d.dump(d.written, input[i:j])
		str.Write(d.buffer)
		if j < len(input) {
			str.WriteRune('\n')
		}
		d.written += uint64(j - i)
	}
	return str.String()
}

func (d *Dumper) Reset() {
	d.written = 0
}

func (d *Dumper) dump(offset uint64, input []byte) {
	d.index = 0
	d.writeOffset(offset)
	d.index += 2 + len(d.delim)
	d.writeInput(input)
	d.index += 2 + len(d.delim)
	d.writeASCII(input)
}

const chars = "0123456789abcdef"

func (d *Dumper) writeOffset(offset uint64) {
	for i := 3; i >= 0; i-- {
		b := (offset >> (i * 8)) & 0xFF
		d.writeByte(chars[b>>4])
		d.writeByte(chars[b&0x0F])
	}
}

func (d *Dumper) writeInput(input []byte) {
	var (
		n = len(input)
		z int
	)
	for i := 0; i < n; i += d.width {
		j := i + d.width
		if j >= n {
			j = n
		}
		var g int
		for k := i; k < j; k++ {
			if g >= d.groups {
				d.seek(1)
				g = 0
				z++
			}
			if !d.bits {
				d.writeByte(chars[input[k]>>4])
				d.writeByte(chars[input[k]&0x0F])
				z += 2
			} else {
				for i := 7; i >= 0; i-- {
					b := bitChar((input[k] >> i) & 0x1)
					d.writeByte(b)
				}
				z += 8
			}
			g++
		}
		if j < n {
			d.writeString(d.padding)
			z += len(d.padding)
		}
	}
	if z < d.size {
		d.writeString(strings.Repeat(" ", d.size-z))
	}
}

func (d *Dumper) writeASCII(input []byte) {
	var z int
	for i := range input {
		if i > 0 && i%d.width == 0 {
			d.seek(1)
			z++
		}
		d.writeByte(byteChar(input[i]))
		z++
	}
	if width := (d.cols * d.width) + (d.cols - 1); z < width {
		d.writeString(strings.Repeat(" ", width-z))
	}
}

func (d *Dumper) seek(n int) {
	d.index += n
}

func (d *Dumper) writeByte(b byte) {
	d.buffer[d.index] = b
	d.index++
}

func (d *Dumper) writeString(str string) {
	d.index += copy(d.buffer[d.index:], []byte(str))
}

func bitChar(b byte) byte {
	if b == 0 {
		return '0'
	}
	return '1'
}

func byteChar(b byte) byte {
	if b >= 32 && b <= 126 {
		return b
	}
	return '.'
}

func (d *Dumper) BlockSize() int {
	return d.cols * d.width
}
