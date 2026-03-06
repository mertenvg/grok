# Grok it like you mean it!

Tired of debugging using `fmt.Println` and `fmt.Printf("%+v", var)`. Use `grok` to print out a pretty formatted view of your variables and what's in them.

## Install:
```sh
go get github.com/mertenvg/grok@latest
```

## Usage:

```go
grok.V(myVar) // or grok.Value(myVar)

// or for customised output

grok.V(myVar, grok.WithMaxDepth(3), grok.WithTabStop(2))
```

The grok package comes with the following customisation options baked in:

### Basic Options

```go
// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option

// WithoutColours disables colouring of output from debug functions. Defaults to `true`
func WithoutColours() Option

// WithMaxDepth sets the maximum recursion depth from debug functions. Defaults to `10`, use `0` for unlimited
func WithMaxDepth(depth int) Option

// WithMaxLength sets the maximum length of string values. Default is `100`, use `0` for unlimited
func WithMaxLength(chars int) Option

// WithTabStop sets the width of a tabstop to the given char count. Defaults to `4`
func WithTabStop(chars int) Option
```

### Advanced Options

```go
// WithContext sets the context for cancellation during deep recursion
func WithContext(ctx context.Context) Option

// WithFilter sets a filter function to conditionally skip fields/values
func WithFilter(filter FilterFunc) Option

// WithErrorHandler sets an error handler for reflection panics and write errors
func WithErrorHandler(handler ErrorHandlerFunc) Option

// WithPrefix adds a prefix to each line of output
func WithPrefix(prefix string) Option

// WithSuffix adds a suffix to the output
func WithSuffix(suffix string) Option

// WithStats enables statistics collection during traversal
func WithStats(stats *Stats) Option
```

### Alternative Output Functions

```go
// S returns the formatted output as a string instead of printing
func S(value any, options ...Option) string

// B returns the formatted output as bytes instead of printing
func B(value any, options ...Option) []byte
```

## Examples

### Basic Usage
```go
package main

import "github.com/mertenvg/grok"

func main() {
    data := map[string]interface{}{
        "name": "John Doe",
        "age": 30,
        "items": []string{"apple", "banana"},
    }

    grok.V(data)
}
```

### Using Context for Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

grok.V(largeStructure, grok.WithContext(ctx))
```

### Filtering Fields
```go
// Skip private fields
filter := func(name string, v reflect.Value) bool {
    return !strings.HasPrefix(name, "_")
}

grok.V(myStruct, grok.WithFilter(filter))
```

### Collecting Statistics
```go
stats := &grok.Stats{}
grok.V(myData, grok.WithStats(stats))
fmt.Printf("Max depth: %d, Types seen: %v\n", stats.MaxDepthReached, stats.TypesSeen)
```

### Getting String Output
```go
output := grok.S(myVar, grok.WithoutColours())
log.Println(output)
```

## Abbreviated types
The following go types have their output abbreviated when nested within other structs. To see their un-abbreviated content simply grok.V that specific value.
```go
time.Time
http.Request
http.Response
```

## Concurrency Safety

The grok package is now fully concurrency-safe:
- Multiple goroutines can safely call `grok.V()` simultaneously to different writers
- Per-writer locking ensures output integrity without serializing all operations
- Context support allows graceful cancellation of long-running traversals

## Got 99 problems and this code is one?

Please create an [issue](https://github.com/mertenvg/grok/issues)
