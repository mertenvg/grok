package grok

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"time"
)

var (
	dumpMutex   = sync.Mutex{}
	grokkerType = reflect.TypeOf((*Grokker)(nil)).Elem()
)

type Grokker interface {
	Grok() string
}

type value struct {
	Name        string
	IsValid     bool
	IsPointer   bool
	IsInterface bool
	Type        string
	RValue      reflect.Value
	RType       reflect.Type
	Elem        any
	Children    []value
}

// V aliases Value
func V(value any, options ...Option) {
	Value(value, options...)
}

// Value prints a more human readable representation of any value
func Value(value any, options ...Option) {
	c := defaults
	for _, o := range options {
		o(&c)
	}

	// only dump one value at a time to avoid overlap
	dumpMutex.Lock()
	defer dumpMutex.Unlock()

	// create a buffer for our output to make it concurrency safe.
	var b bytes.Buffer

	// dump it like you mean it
	dump("value", reflect.ValueOf(value), writer(&b), colourizer(c.colour), indenter(c.tabstop), c.depth, c.maxDepth, c.maxLength)

	// Write the contents of the buffer to the configured writer.
	c.writer.Write(b.Bytes())
}

func dump(name string, v reflect.Value, write Writer, colour Colourizer, indent Indenter, depth, maxDepth, maxLength int) {
	val := value{
		Name:   name,
		RValue: v,
	}

	if !v.IsValid() {
		val.Elem = "<invalid>"
		write(indent("<invalid>\n", depth))
		return
	}
	t := v.Type()

	tn := ""
	switch t.Kind() {
	case reflect.Interface:
		name := t.Name()
		tn = colour("any", colourBlue) + colour(name, colourBlue)
		if !v.IsNil() {
			v = v.Elem()
			t = v.Type()
		}
		if t.Kind() == reflect.Ptr {
			v = v.Elem()
			t = v.Type()
		}
		if t.Name() != name {
			tn = colour(t.String(), colourBlue) + colour(" as ", colourRed) + tn
		}
	case reflect.Ptr:
		v = v.Elem()
		t = t.Elem()
		tn = colour("*", colourRed) + colour(t.String(), colourBlue)
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

	if len(name) > 0 {
		write(indent("%s %s = ", depth), colour(name, colourYellow), tn)
	} else {
		write(indent("", depth))
	}

	switch v.Kind() {
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
		depth = depth + 1
		if maxDepth > 0 && depth >= maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), depth))
		} else {
			for i := 0; i < v.Len(); i++ {
				dump(colour(fmt.Sprintf("%d", i), colourRed), v.Index(i), write, colour, indent, depth, maxDepth, maxLength)
			}
		}
		depth = depth - 1
		write(indent("]", depth))
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
			depth = depth + 1
			if maxDepth > 0 && depth >= maxDepth {
				write(indent(colour("... max depth reached\n", colourGrey), depth))
			} else {
				keys := v.MapKeys()
				sort.Sort(byValue(keys))
				for _, k := range v.MapKeys() {
					dump(fmt.Sprintf("%v", k), v.MapIndex(k), write, colour, indent, depth, maxDepth, maxLength)
				}
			}
			depth = depth - 1
			write(indent("]", depth))
		}
	case reflect.String:
		s := v.String()
		slen := len(s)
		if maxLength > 0 && slen > maxLength {
			s = fmt.Sprintf("%s...", string([]byte(s)[0:maxLength]))
		}
		write(colour("%q ", colourGreen), s)
		write(colour("%d", colourGrey), slen)
	case reflect.Struct:
		if v.NumField() == 0 {
			write("{}\n")
			return
		}
		write("{\n")
		depth = depth + 1
		if maxDepth > 0 && depth >= maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), depth))
		} else {
			switch {
			case !v.CanInterface():
				write(indent(colour(fmt.Sprintf("... ???\n"), colourGrey), depth))
			case v.Type().Implements(grokkerType):
				o := v.Interface().(Grokker)
				write(indent(colour(fmt.Sprintf("... %s\n", o.Grok()), colourGrey), depth))
			case depth > 1 && t.String() == "time.Time":
				write(indent(colour(fmt.Sprintf("... %v\n", v), colourGrey), depth))
			case depth > 1 && t.String() == "time.Location":
				s := "<nil>"
				if v.CanAddr() && v.CanInterface() {
					s = v.Addr().Interface().(*time.Location).String()
				}
				write(indent(colour(fmt.Sprintf("... %v\n", s), colourGrey), depth))
			case depth > 1 && t.String() == "http.Request":
				o := v.Interface().(http.Request)
				write(indent(colour(fmt.Sprintf("... %s %s %d\n", coalesce(o.Method, "GET"), coalesce(o.RequestURI, "<request-uri>"), o.ContentLength), colourGrey), depth))
			case depth > 1 && t.String() == "http.Response":
				o := v.Interface().(http.Response)
				write(indent(colour(fmt.Sprintf("... %s %d\n", coalesce(o.Status, "<status-code> <status>"), o.ContentLength), colourGrey), depth))
			default:
				for i := 0; i < v.NumField(); i++ {
					dump(t.Field(i).Name, v.Field(i), write, colour, indent, depth, maxDepth, maxLength)
				}
			}

		}
		depth = depth - 1
		write(indent("}", depth))
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
