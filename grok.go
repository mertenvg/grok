package grok

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
)

var (
	colourReset  = "\x1B[0m"
	colourRed    = "\x1B[38;5;124m"
	colourYellow = "\x1B[38;5;208m"
	colourBlue   = "\x1B[38;5;33m"
	colourGrey   = "\x1B[38;5;144m"
	colourGreen  = "\x1B[38;5;34m"
	colourGold   = "\x1B[38;5;3m"
	writerMutexes = sync.Map{} // map[io.Writer]*sync.Mutex
)

type Writer func(string, ...any)

type Indenter func(string, int) string

type Colourizer func(str string, colour string) string

type Option func(c *Conf)

type FilterFunc func(name string, v reflect.Value) bool

type ErrorHandlerFunc func(err error)

type Stats struct {
	MaxDepthReached int
	TypesSeen       map[string]int
	FieldsTraversed int
}

type Conf struct {
	depth        int
	writer       io.Writer
	tabstop      int
	colour       bool
	maxDepth     int
	maxLength    int
	ctx          context.Context
	filter       FilterFunc
	errorHandler ErrorHandlerFunc
	prefix       string
	suffix       string
	collectStats bool
	stats        *Stats
}

func getDefaults() Conf {
	return Conf{
		depth:     0,
		writer:    os.Stdout,
		tabstop:   4,
		colour:    true,
		maxDepth:  10,
		maxLength: 100,
		ctx:       context.Background(),
	}
}

func writer(w io.Writer, errorHandler ErrorHandlerFunc) Writer {
	return Writer(func(format string, params ...any) {
		_, err := w.Write([]byte(fmt.Sprintf(format, params...)))
		if err != nil && errorHandler != nil {
			errorHandler(fmt.Errorf("write error: %w", err))
		}
	})
}

func indenter(tabstop int) Indenter {
	return Indenter(func(v string, depth int) string {
		return strings.Repeat(" ", depth*tabstop) + v
	})
}

func colourizer(colourize bool) Colourizer {
	return Colourizer(func(str string, colour string) string {
		if colourize {
			return colour + str + colourReset
		}
		return str
	})
}

// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option {
	return func(c *Conf) {
		c.writer = w
	}
}

// WithoutColours disables colouring of output from debug functions. Defaults to `true`
func WithoutColours() Option {
	return func(c *Conf) {
		c.colour = false
	}
}

// WithMaxDepth sets the maximum recursion depth from debug functions. Defaults to `10`, use `0` for unlimited
func WithMaxDepth(depth int) Option {
	return func(c *Conf) {
		c.maxDepth = depth
	}
}

// WithMaxLength sets the maximum length of string byValue. Default is `100`, use `0` for unlimited
func WithMaxLength(chars int) Option {
	return func(c *Conf) {
		c.maxLength = chars
	}
}

// WithTabStop sets the width of a tabstop to the given char count. Defaults to `4`
func WithTabStop(chars int) Option {
	return func(c *Conf) {
		c.tabstop = chars
	}
}

// WithContext sets the context for cancellation during deep recursion
func WithContext(ctx context.Context) Option {
	return func(c *Conf) {
		c.ctx = ctx
	}
}

// WithFilter sets a filter function to conditionally skip fields/values
func WithFilter(filter FilterFunc) Option {
	return func(c *Conf) {
		c.filter = filter
	}
}

// WithErrorHandler sets an error handler for reflection panics
func WithErrorHandler(handler ErrorHandlerFunc) Option {
	return func(c *Conf) {
		c.errorHandler = handler
	}
}

// WithPrefix adds a prefix to each line of output
func WithPrefix(prefix string) Option {
	return func(c *Conf) {
		c.prefix = prefix
	}
}

// WithSuffix adds a suffix to the output
func WithSuffix(suffix string) Option {
	return func(c *Conf) {
		c.suffix = suffix
	}
}

// WithStats enables statistics collection during traversal
func WithStats(stats *Stats) Option {
	return func(c *Conf) {
		c.collectStats = true
		c.stats = stats
		if c.stats.TypesSeen == nil {
			c.stats.TypesSeen = make(map[string]int)
		}
	}
}

// getMutexForWriter returns a mutex for the given writer
func getMutexForWriter(w io.Writer) *sync.Mutex {
	if m, ok := writerMutexes.Load(w); ok {
		return m.(*sync.Mutex)
	}
	m := &sync.Mutex{}
	actual, _ := writerMutexes.LoadOrStore(w, m)
	return actual.(*sync.Mutex)
}
