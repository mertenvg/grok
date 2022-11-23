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
 
```go
// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option
```
```go
// WithoutColours disables colouring of output from debug functions. Defaults to `true`
func WithoutColours() Option
```
```go
// WithMaxDepth sets the maximum recursion depth from debug functions. Defaults to `10`, use `0` for unlimited
func WithMaxDepth(depth int) Option 
```
```go
// WithMaxLength sets the maximum length of string values. Default is `100`, use `0` for unlimited
func WithMaxLength(chars int) Option
```
```go
// WithTabStop sets the width of a tabstop to the given char count. Defaults to `4`
func WithTabStop(chars int) Option
```

## Abbreviated types
The following go types have their output abbreviated when nested within other structs. To see their un-abbreviated content simply grok.V that specific value.
```go
time.Time
http.Request
http.Response
```

## Got 99 problems and this code is one?

Please create an [issue](https://github.com/mertenvg/grok/issues)
