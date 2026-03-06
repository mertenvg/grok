package grok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"time"
)

var (
	grokkerType = reflect.TypeOf((*Grokker)(nil)).Elem()
)

type Grokker interface {
	Grok() string
}

// V aliases Value
func V(value any, options ...Option) {
	Value(value, options...)
}

// Value prints a more human readable representation of any value
func Value(value any, options ...Option) {
	c := getDefaults()
	for _, o := range options {
		o(&c)
	}

	// Get a mutex specific to this writer to allow concurrent dumps to different writers
	mutex := getMutexForWriter(c.writer)
	mutex.Lock()
	defer mutex.Unlock()

	// create a buffer for our output to make it concurrency safe.
	var b bytes.Buffer

	// Add prefix if configured
	if c.prefix != "" {
		b.WriteString(c.prefix)
	}

	// dump it like you mean it
	dump("value", reflect.ValueOf(value), writer(&b, c.errorHandler), colourizer(c.colour), indenter(c.tabstop), &c)

	// Add suffix if configured
	if c.suffix != "" {
		b.WriteString(c.suffix)
	}

	// Write the contents of the buffer to the configured writer.
	_, err := c.writer.Write(b.Bytes())
	if err != nil && c.errorHandler != nil {
		c.errorHandler(fmt.Errorf("write error: %w", err))
	}
}

// S returns the formatted output as a string
func S(value any, options ...Option) string {
	var b bytes.Buffer
	options = append(options, WithWriter(&b))
	Value(value, options...)
	return b.String()
}

// B returns the formatted output as bytes
func B(value any, options ...Option) []byte {
	var b bytes.Buffer
	options = append(options, WithWriter(&b))
	Value(value, options...)
	return b.Bytes()
}

func dump(name string, v reflect.Value, write Writer, colour Colourizer, indent Indenter, c *Conf) {
	// Check context cancellation
	if c.ctx != nil {
		select {
		case <-c.ctx.Done():
			write(indent(colour("... context cancelled\n", colourGrey), c.depth))
			return
		default:
		}
	}

	// Check filter
	if c.filter != nil && !c.filter(name, v) {
		return
	}

	// Collect stats
	if c.collectStats && c.stats != nil {
		c.stats.FieldsTraversed++
		if c.depth > c.stats.MaxDepthReached {
			c.stats.MaxDepthReached = c.depth
		}
	}

	// Recover from panics during reflection
	defer func() {
		if r := recover(); r != nil {
			if c.errorHandler != nil {
				c.errorHandler(fmt.Errorf("panic during reflection: %v", r))
			}
			write(indent(colour(fmt.Sprintf("... panic: %v\n", r), colourRed), c.depth))
		}
	}()

	if !v.IsValid() {
		write(indent("<invalid>\n", c.depth))
		return
	}
	t := v.Type()

	// Collect type statistics
	if c.collectStats && c.stats != nil {
		c.stats.TypesSeen[t.String()]++
	}

	tn := formatTypeName(t, v, colour)

	if len(name) > 0 {
		write(indent("%s %s = ", c.depth), colour(name, colourYellow), tn)
	} else {
		write(indent("", c.depth))
	}

	formatValue(v, t, name, write, colour, indent, c)
}

func formatTypeName(t reflect.Type, v reflect.Value, colour Colourizer) string {
	tn := ""
	switch t.Kind() {
	case reflect.Interface:
		name := t.Name()
		tn = colour("any", colourBlue) + colour(name, colourBlue)
		if !v.IsNil() {
			newV := v.Elem()
			newT := newV.Type()
			if newT.Kind() == reflect.Ptr {
				newV = newV.Elem()
				newT = newV.Type()
			}
			if newT.Name() != name {
				tn = colour(newT.String(), colourBlue) + colour(" as ", colourRed) + tn
			}
		}
	case reflect.Ptr:
		tn = colour("*", colourRed) + colour(t.Elem().String(), colourBlue)
	case reflect.Slice:
		tn = colour("[]", colourRed)
		switch t.Elem().Kind() {
		case reflect.Interface:
			tn = tn + colour(t.Elem().String(), colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Elem().Elem().String(), colourBlue)
		default:
			tn = tn + colour(t.Elem().String(), colourBlue)
		}
	case reflect.Map:
		tn = colour("map[", colourRed)
		switch t.Key().Kind() {
		case reflect.Interface:
			tn = tn + colour(t.Key().String(), colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Key().Elem().String(), colourBlue)
		default:
			tn = tn + colour(t.Key().String(), colourBlue)
		}
		tn = tn + colour("]", colourRed)
		switch t.Elem().Kind() {
		case reflect.Interface:
			tn = tn + colour("any", colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Elem().Elem().String(), colourBlue)
		default:
			tn = tn + colour(t.Elem().String(), colourBlue)
		}
	case reflect.Chan:
		tn = colour(t.ChanDir().String(), colourRed)
		tn = tn + " " + colour(t.Elem().String(), colourBlue)
	case reflect.Func:
		tn = colour("func", colourRed)
	case reflect.UnsafePointer:
		tn = colour("unsafe*", colourRed) + colour(t.String(), colourBlue)
	default:
		tn = colour(t.String(), colourBlue)
	}
	return tn
}

func formatValue(v reflect.Value, t reflect.Type, name string, write Writer, colour Colourizer, indent Indenter, c *Conf) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			// Dereference and continue dumping
			dump("", v.Elem(), write, colour, indent, c)
			return
		}
	case reflect.Bool:
		write(colour("%v", colourGreen), v.Bool())
	case reflect.Uintptr:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		write(colour("%v", colourGreen), v)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Len() == 0 {
			write("[]\n")
			return
		}
		write("[\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), c.depth))
		} else {
			for i := 0; i < v.Len(); i++ {
				dump(colour(fmt.Sprintf("%d", i), colourRed), v.Index(i), write, colour, indent, c)
			}
		}
		c.depth = c.depth - 1
		write(indent("]", c.depth))
	case reflect.Chan:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	case reflect.Func:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	case reflect.Map:
		if !v.IsValid() {
			write(colour("<nil>", colourGrey))
		} else {
			if v.Len() == 0 {
				write("[]\n")
				return
			}
			write("[\n")
			c.depth = c.depth + 1
			if c.maxDepth > 0 && c.depth >= c.maxDepth {
				write(indent(colour("... max depth reached\n", colourGrey), c.depth))
			} else {
				keys := v.MapKeys()
				sort.Sort(byValue(keys))
				for _, k := range keys {
					dump(fmt.Sprintf("%v", k), v.MapIndex(k), write, colour, indent, c)
				}
			}
			c.depth = c.depth - 1
			write(indent("]", c.depth))
		}
	case reflect.String:
		s := v.String()
		slen := len(s)
		if c.maxLength > 0 && slen > c.maxLength {
			s = fmt.Sprintf("%s...", string([]byte(s)[0:c.maxLength]))
		}
		write(colour("%q ", colourGreen), s)
		write(colour("%d", colourGrey), slen)
	case reflect.Struct:
		if v.NumField() == 0 {
			write("{}\n")
			return
		}
		write("{\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), c.depth))
		} else {
			switch {
			case !v.CanInterface():
				write(indent(colour(fmt.Sprintf("... ???\n"), colourGrey), c.depth))
			case v.Type().Implements(grokkerType):
				o := v.Interface().(Grokker)
				write(indent(colour(fmt.Sprintf("... %s\n", o.Grok()), colourGrey), c.depth))
			case t.String() == "json.RawMessage":
				o := v.Interface().(json.RawMessage)
				write(indent(colour(fmt.Sprintf("... %s\n", string(o)), colourGrey), c.depth))
			case c.depth > 1 && t.String() == "time.Time":
				write(indent(colour(fmt.Sprintf("... %v\n", v), colourGrey), c.depth))
			case c.depth > 1 && t.String() == "time.Location":
				s := "<nil>"
				if v.CanAddr() && v.CanInterface() {
					s = v.Addr().Interface().(*time.Location).String()
				}
				write(indent(colour(fmt.Sprintf("... %v\n", s), colourGrey), c.depth))
			case c.depth > 1 && t.String() == "http.Request":
				o := v.Interface().(http.Request)
				write(indent(colour(fmt.Sprintf("... %s %s %d\n", coalesce(o.Method, "GET"), coalesce(o.RequestURI, "<request-uri>"), o.ContentLength), colourGrey), c.depth))
			case c.depth > 1 && t.String() == "http.Response":
				o := v.Interface().(http.Response)
				write(indent(colour(fmt.Sprintf("... %s %d\n", coalesce(o.Status, "<status-code> <status>"), o.ContentLength), colourGrey), c.depth))
			default:
				for i := 0; i < v.NumField(); i++ {
					dump(t.Field(i).Name, v.Field(i), write, colour, indent, c)
				}
			}

		}
		c.depth = c.depth - 1
		write(indent("}", c.depth))
	case reflect.UnsafePointer:
		write(colour("%v", colourGreen), v)
	case reflect.Invalid:
		write(colour("<nil>", colourGrey))
	case reflect.Interface:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	default:
		write(colour("??? %s", colourRed), v.Kind().String())
	}
	write("\n")
}

type byValue []reflect.Value

// Len implements sort.Interface
func (s byValue) Len() int {
	return len(s)
}

// Swap implements sort.Interface
func (s byValue) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface
func (s byValue) Less(i, j int) bool {
	return fmt.Sprintf("%v", s[i]) < fmt.Sprintf("%v", s[j])
}

// coalesce returns the first non-empty value from the given arguments
func coalesce(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
