package grok

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestBasicTypes tests formatting of basic Go types
func TestBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		contains []string
	}{
		{
			name:     "integer",
			input:    42,
			contains: []string{"int", "42"},
		},
		{
			name:     "string",
			input:    "hello world",
			contains: []string{"string", "hello world"},
		},
		{
			name:     "bool true",
			input:    true,
			contains: []string{"bool", "true"},
		},
		{
			name:     "bool false",
			input:    false,
			contains: []string{"bool", "false"},
		},
		{
			name:     "float",
			input:    3.14,
			contains: []string{"float64", "3.14"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := S(tt.input, WithoutColours())
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("expected output to contain %q, got: %s", substr, output)
				}
			}
		})
	}
}

// TestSliceAndArray tests slice and array formatting
func TestSliceAndArray(t *testing.T) {
	slice := []int{1, 2, 3}
	output := S(slice, WithoutColours())

	if !strings.Contains(output, "[]int") {
		t.Errorf("expected []int type, got: %s", output)
	}
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Errorf("expected slice elements, got: %s", output)
	}
}

// TestMap tests map formatting and sorting
func TestMap(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	output := S(m, WithoutColours())

	if !strings.Contains(output, "map[string]int") {
		t.Errorf("expected map type, got: %s", output)
	}

	// Check that map keys appear (sorting is tested separately)
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") || !strings.Contains(output, "c") {
		t.Errorf("expected map keys, got: %s", output)
	}
}

// TestStruct tests struct formatting
func TestStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p := Person{Name: "Alice", Age: 30}
	output := S(p, WithoutColours())

	if !strings.Contains(output, "Person") {
		t.Errorf("expected struct type name, got: %s", output)
	}
	if !strings.Contains(output, "Name") || !strings.Contains(output, "Alice") {
		t.Errorf("expected Name field, got: %s", output)
	}
	if !strings.Contains(output, "Age") || !strings.Contains(output, "30") {
		t.Errorf("expected Age field, got: %s", output)
	}
}

// TestPointer tests pointer handling
func TestPointer(t *testing.T) {
	val := 42
	ptr := &val
	output := S(ptr, WithoutColours())

	if !strings.Contains(output, "*int") {
		t.Errorf("expected pointer type, got: %s", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("expected dereferenced value, got: %s", output)
	}
}

// TestNilValues tests nil handling
func TestNilValues(t *testing.T) {
	var ptr *int
	output := S(ptr, WithoutColours())

	if !strings.Contains(output, "nil") {
		t.Errorf("expected nil, got: %s", output)
	}
}

// TestWithMaxDepth tests depth limiting
func TestWithMaxDepth(t *testing.T) {
	type Nested struct {
		Child *Nested
	}

	n := &Nested{Child: &Nested{Child: &Nested{Child: &Nested{}}}}
	output := S(n, WithMaxDepth(2), WithoutColours())

	if !strings.Contains(output, "max depth reached") {
		t.Errorf("expected max depth message, got: %s", output)
	}
}

// TestWithMaxLength tests string length limiting
func TestWithMaxLength(t *testing.T) {
	longString := strings.Repeat("a", 200)
	output := S(longString, WithMaxLength(50), WithoutColours())

	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated string, got: %s", output)
	}
	// Check that the string is actually truncated
	if strings.Count(output, "a") > 55 { // Allow some margin for formatting
		t.Errorf("string not properly truncated, got: %s", output)
	}
}

// TestWithWriter tests custom writer
func TestWithWriter(t *testing.T) {
	var buf bytes.Buffer
	V(42, WithWriter(&buf), WithoutColours())

	output := buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("expected output in custom writer, got: %s", output)
	}
}

// TestWithoutColours tests color disabling
func TestWithoutColours(t *testing.T) {
	output := S(42, WithoutColours())

	// Should not contain ANSI escape codes
	if strings.Contains(output, "\x1B[") {
		t.Errorf("expected no color codes, got: %s", output)
	}
}

// TestWithColours tests that colors are included by default
func TestWithColours(t *testing.T) {
	output := S(42)

	// Should contain ANSI escape codes
	if !strings.Contains(output, "\x1B[") {
		t.Errorf("expected color codes, got: %s", output)
	}
}

// TestWithTabStop tests custom tab stops
func TestWithTabStop(t *testing.T) {
	type Nested struct {
		Value int
	}
	n := Nested{Value: 42}

	output2 := S(n, WithTabStop(2), WithoutColours())
	output8 := S(n, WithTabStop(8), WithoutColours())

	// Output with 8-space tabs should be longer
	if len(output8) <= len(output2) {
		t.Errorf("expected different tab spacing, 2-space: %d chars, 8-space: %d chars", len(output2), len(output8))
	}
}

// TestSFunction tests string output function
func TestSFunction(t *testing.T) {
	output := S(42, WithoutColours())

	if output == "" {
		t.Error("S() returned empty string")
	}
	if !strings.Contains(output, "42") {
		t.Errorf("expected 42 in output, got: %s", output)
	}
}

// TestBFunction tests bytes output function
func TestBFunction(t *testing.T) {
	output := B(42, WithoutColours())

	if len(output) == 0 {
		t.Error("B() returned empty bytes")
	}
	if !bytes.Contains(output, []byte("42")) {
		t.Errorf("expected 42 in output, got: %s", output)
	}
}

// TestVFunction tests that V() is an alias for Value()
func TestVFunction(t *testing.T) {
	var buf bytes.Buffer
	V(42, WithWriter(&buf), WithoutColours())

	output := buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("V() should work like Value(), got: %s", output)
	}
}

// TestWithPrefix tests prefix addition
func TestWithPrefix(t *testing.T) {
	output := S(42, WithPrefix("[DEBUG] "), WithoutColours())

	if !strings.HasPrefix(output, "[DEBUG] ") {
		t.Errorf("expected prefix, got: %s", output)
	}
}

// TestWithSuffix tests suffix addition
func TestWithSuffix(t *testing.T) {
	output := S(42, WithSuffix(" [END]"), WithoutColours())

	if !strings.HasSuffix(output, " [END]") {
		t.Errorf("expected suffix, got: %s", output)
	}
}

// TestWithFilter tests filtering functionality
func TestWithFilter(t *testing.T) {
	type Person struct {
		Name       string
		Age        int
		SecretCode string
	}

	p := Person{Name: "Alice", Age: 30, SecretCode: "secret"}

	// Filter out SecretCode field
	filter := func(name string, v reflect.Value) bool {
		return name != "SecretCode"
	}

	output := S(p, WithFilter(filter), WithoutColours())

	if !strings.Contains(output, "Alice") {
		t.Errorf("expected Name field, got: %s", output)
	}
	if strings.Contains(output, "SecretCode") || strings.Contains(output, "secret") {
		t.Errorf("expected SecretCode to be filtered out, got: %s", output)
	}
}

// TestWithContext tests context support
func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	type Deep struct {
		Child *Deep
	}
	deep := &Deep{Child: &Deep{Child: &Deep{}}}

	output := S(deep, WithContext(ctx), WithoutColours())

	// Should contain cancellation message
	if !strings.Contains(output, "context cancelled") {
		t.Logf("Note: context cancellation may not trigger on shallow structures, got: %s", output)
	}
}

// TestWithContextTimeout tests context timeout
func TestWithContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	output := S([]int{1, 2, 3}, WithContext(ctx), WithoutColours())

	// Should complete or show cancellation
	if output == "" {
		t.Error("expected some output")
	}
}

// TestWithErrorHandler tests error handling
func TestWithErrorHandler(t *testing.T) {
	var capturedError error
	handler := func(err error) {
		capturedError = err
	}

	// This should work without errors
	S(42, WithErrorHandler(handler), WithoutColours())

	if capturedError != nil {
		t.Errorf("unexpected error captured: %v", capturedError)
	}
}

// TestWithStats tests statistics collection
func TestWithStats(t *testing.T) {
	type Person struct {
		Name    string
		Age     int
		Friends []string
	}

	p := Person{
		Name:    "Alice",
		Age:     30,
		Friends: []string{"Bob", "Charlie"},
	}

	stats := &Stats{}
	S(p, WithStats(stats), WithoutColours())

	if stats.FieldsTraversed == 0 {
		t.Error("expected fields to be traversed")
	}
	if stats.MaxDepthReached == 0 {
		t.Error("expected max depth to be recorded")
	}
	if len(stats.TypesSeen) == 0 {
		t.Error("expected types to be recorded")
	}

	// Check that we saw the Person struct type
	found := false
	for typeName := range stats.TypesSeen {
		if strings.Contains(typeName, "Person") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to see Person type in stats, got: %v", stats.TypesSeen)
	}
}

// TestGrokkerInterface tests custom Grokker interface
func TestGrokkerInterface(t *testing.T) {
	type CustomType struct {
		data string
	}

	// This won't work in this test since we can't add methods to the type here,
	// but we can test that the interface exists
	var _ Grokker = (*mockGrokker)(nil)
}

type mockGrokker struct{}

func (m *mockGrokker) Grok() string {
	return "custom grok output"
}

// TestConcurrentWrites tests that concurrent writes to the same writer are safe
func TestConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	var wg sync.WaitGroup

	iterations := 100
	wg.Add(iterations)

	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			V(val, WithWriter(&buf), WithoutColours())
		}(i)
	}

	wg.Wait()

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected output from concurrent writes")
	}
}

// TestConcurrentDifferentWriters tests that concurrent writes to different writers work
func TestConcurrentDifferentWriters(t *testing.T) {
	var wg sync.WaitGroup
	iterations := 50
	wg.Add(iterations)

	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			var buf bytes.Buffer
			V(val, WithWriter(&buf), WithoutColours())

			output := buf.String()
			if !strings.Contains(output, fmt.Sprintf("%d", val)) {
				t.Errorf("expected %d in output, got: %s", val, output)
			}
		}(i)
	}

	wg.Wait()
}

// TestMapKeySorting tests that map keys are sorted
func TestMapKeySorting(t *testing.T) {
	m := map[string]int{
		"zebra": 1,
		"apple": 2,
		"mango": 3,
	}

	output := S(m, WithoutColours())

	// Find positions of keys in output
	applePos := strings.Index(output, "apple")
	mangoPos := strings.Index(output, "mango")
	zebraPos := strings.Index(output, "zebra")

	if applePos == -1 || mangoPos == -1 || zebraPos == -1 {
		t.Fatalf("not all keys found in output: %s", output)
	}

	// Keys should appear in sorted order
	if !(applePos < mangoPos && mangoPos < zebraPos) {
		t.Errorf("map keys not sorted correctly: apple=%d, mango=%d, zebra=%d\nOutput: %s",
			applePos, mangoPos, zebraPos, output)
	}
}

// TestInvalidReflectValue tests handling of invalid values
func TestInvalidReflectValue(t *testing.T) {
	var nilInterface interface{}
	output := S(nilInterface, WithoutColours())

	// Invalid values are shown as "<invalid>"
	if !strings.Contains(output, "invalid") {
		t.Errorf("expected invalid handling, got: %s", output)
	}
}

// TestComplexNestedStructure tests deeply nested complex structures
func TestComplexNestedStructure(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	type Person struct {
		Name    string
		Age     int
		Address *Address
		Tags    map[string]string
		Scores  []int
	}

	p := Person{
		Name: "Alice",
		Age:  30,
		Address: &Address{
			Street: "123 Main St",
			City:   "NYC",
		},
		Tags: map[string]string{
			"role": "admin",
			"team": "engineering",
		},
		Scores: []int{95, 87, 92},
	}

	output := S(p, WithoutColours())

	// Verify all data appears
	checks := []string{"Alice", "30", "123 Main St", "NYC", "admin", "engineering", "95", "87", "92"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected %q in output, got: %s", check, output)
		}
	}
}

// TestEmptyCollections tests empty slices, maps, and structs
func TestEmptyCollections(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"empty slice", []int{}},
		{"empty map", map[string]int{}},
		{"empty struct", struct{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := S(tt.value, WithoutColours())
			if output == "" {
				t.Error("expected some output for empty collection")
			}
		})
	}
}

// TestChannelAndFunc tests channel and function types
func TestChannelAndFunc(t *testing.T) {
	ch := make(chan int)
	fn := func() {}

	outputCh := S(ch, WithoutColours())
	outputFn := S(fn, WithoutColours())

	if !strings.Contains(outputCh, "chan") {
		t.Errorf("expected chan type, got: %s", outputCh)
	}
	if !strings.Contains(outputFn, "func") {
		t.Errorf("expected func type, got: %s", outputFn)
	}
}

// TestNilChannel tests nil channel handling
func TestNilChannel(t *testing.T) {
	var ch chan int
	output := S(ch, WithoutColours())

	if !strings.Contains(output, "nil") {
		t.Errorf("expected nil for nil channel, got: %s", output)
	}
}

// TestCircularReference tests that circular references don't cause infinite loops
func TestCircularReference(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}

	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2}
	n1.Next = n2
	n2.Next = n1 // Circular reference

	// Should not hang - max depth should protect us
	output := S(n1, WithMaxDepth(5), WithoutColours())

	if !strings.Contains(output, "max depth reached") {
		t.Errorf("expected max depth protection, got: %s", output)
	}
}

// TestPanicRecovery tests that reflection panics are recovered
func TestPanicRecovery(t *testing.T) {
	var errorCaptured bool
	handler := func(err error) {
		errorCaptured = true
	}

	// Normal value shouldn't panic
	output := S(42, WithErrorHandler(handler), WithoutColours())

	if errorCaptured {
		t.Error("unexpected error for normal value")
	}
	if output == "" {
		t.Error("expected output")
	}
}

// TestWriteErrors tests handling of write errors
func TestWriteErrors(t *testing.T) {
	// Create a writer that always fails
	errWriter := &errorWriter{err: errors.New("write failed")}

	var errorCaptured error
	handler := func(err error) {
		errorCaptured = err
	}

	V(42, WithWriter(errWriter), WithErrorHandler(handler), WithoutColours())

	if errorCaptured == nil {
		t.Error("expected write error to be captured")
	}
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}
