package grok

import (
	"bytes"
	"reflect"
	"testing"
)

// BenchmarkSimpleInt benchmarks formatting a simple integer
func BenchmarkSimpleInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		S(42, WithoutColours())
	}
}

// BenchmarkSimpleString benchmarks formatting a simple string
func BenchmarkSimpleString(b *testing.B) {
	str := "hello world"
	for i := 0; i < b.N; i++ {
		S(str, WithoutColours())
	}
}

// BenchmarkSlice benchmarks formatting a slice
func BenchmarkSlice(b *testing.B) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		S(slice, WithoutColours())
	}
}

// BenchmarkMap benchmarks formatting a map
func BenchmarkMap(b *testing.B) {
	m := map[string]int{
		"a": 1, "b": 2, "c": 3, "d": 4, "e": 5,
	}
	for i := 0; i < b.N; i++ {
		S(m, WithoutColours())
	}
}

// BenchmarkStruct benchmarks formatting a struct
func BenchmarkStruct(b *testing.B) {
	type Person struct {
		Name  string
		Age   int
		Email string
	}
	p := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}

	for i := 0; i < b.N; i++ {
		S(p, WithoutColours())
	}
}

// BenchmarkComplexStruct benchmarks formatting a complex nested structure
func BenchmarkComplexStruct(b *testing.B) {
	type Address struct {
		Street  string
		City    string
		ZipCode string
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
			Street:  "123 Main St",
			City:    "NYC",
			ZipCode: "10001",
		},
		Tags: map[string]string{
			"role":       "admin",
			"team":       "engineering",
			"department": "tech",
		},
		Scores: []int{95, 87, 92, 88, 91},
	}

	for i := 0; i < b.N; i++ {
		S(p, WithoutColours())
	}
}

// BenchmarkDeepNesting benchmarks deeply nested structures
func BenchmarkDeepNesting(b *testing.B) {
	type Deep struct {
		Value int
		Child *Deep
	}

	// Create a structure 10 levels deep
	var createDeep func(depth int) *Deep
	createDeep = func(depth int) *Deep {
		if depth == 0 {
			return nil
		}
		return &Deep{
			Value: depth,
			Child: createDeep(depth - 1),
		}
	}

	deep := createDeep(10)

	for i := 0; i < b.N; i++ {
		S(deep, WithoutColours())
	}
}

// BenchmarkWithColors benchmarks with color output
func BenchmarkWithColors(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}
	p := Person{Name: "Alice", Age: 30}

	for i := 0; i < b.N; i++ {
		S(p) // Colors enabled by default
	}
}

// BenchmarkWithoutColors benchmarks without color output
func BenchmarkWithoutColors(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}
	p := Person{Name: "Alice", Age: 30}

	for i := 0; i < b.N; i++ {
		S(p, WithoutColours())
	}
}

// BenchmarkConcurrentWrites benchmarks concurrent writes to different writers
func BenchmarkConcurrentWrites(b *testing.B) {
	type Person struct {
		Name string
		Age  int
	}
	p := Person{Name: "Alice", Age: 30}

	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			V(p, WithWriter(&buf), WithoutColours())
		}
	})
}

// BenchmarkSFunction benchmarks the S() string return function
func BenchmarkSFunction(b *testing.B) {
	data := map[string]interface{}{
		"name": "test",
		"age":  42,
		"list": []int{1, 2, 3},
	}

	for i := 0; i < b.N; i++ {
		_ = S(data, WithoutColours())
	}
}

// BenchmarkBFunction benchmarks the B() bytes return function
func BenchmarkBFunction(b *testing.B) {
	data := map[string]interface{}{
		"name": "test",
		"age":  42,
		"list": []int{1, 2, 3},
	}

	for i := 0; i < b.N; i++ {
		_ = B(data, WithoutColours())
	}
}

// BenchmarkWithFilter benchmarks filtering functionality
func BenchmarkWithFilter(b *testing.B) {
	type Person struct {
		Name       string
		Age        int
		Email      string
		SecretCode string
	}
	p := Person{Name: "Alice", Age: 30, Email: "alice@example.com", SecretCode: "secret"}

	filter := func(name string, v reflect.Value) bool {
		return name != "SecretCode"
	}

	for i := 0; i < b.N; i++ {
		S(p, WithFilter(filter), WithoutColours())
	}
}

// BenchmarkWithStats benchmarks statistics collection
func BenchmarkWithStats(b *testing.B) {
	type Person struct {
		Name   string
		Age    int
		Scores []int
	}
	p := Person{Name: "Alice", Age: 30, Scores: []int{95, 87, 92}}

	for i := 0; i < b.N; i++ {
		stats := &Stats{}
		S(p, WithStats(stats), WithoutColours())
	}
}

// BenchmarkLargeSlice benchmarks a large slice
func BenchmarkLargeSlice(b *testing.B) {
	slice := make([]int, 100)
	for i := range slice {
		slice[i] = i
	}

	for i := 0; i < b.N; i++ {
		S(slice, WithoutColours())
	}
}

// BenchmarkLargeMap benchmarks a large map
func BenchmarkLargeMap(b *testing.B) {
	m := make(map[string]int, 100)
	for i := 0; i < 100; i++ {
		m[string(rune('a'+i%26))+string(rune('0'+i))] = i
	}

	for i := 0; i < b.N; i++ {
		S(m, WithoutColours())
	}
}

// BenchmarkMaxDepthLimit benchmarks with depth limiting
func BenchmarkMaxDepthLimit(b *testing.B) {
	type Deep struct {
		Value int
		Child *Deep
	}

	var createDeep func(depth int) *Deep
	createDeep = func(depth int) *Deep {
		if depth == 0 {
			return nil
		}
		return &Deep{
			Value: depth,
			Child: createDeep(depth - 1),
		}
	}

	deep := createDeep(20)

	for i := 0; i < b.N; i++ {
		S(deep, WithMaxDepth(5), WithoutColours())
	}
}

// BenchmarkMaxLengthLimit benchmarks with string length limiting
func BenchmarkMaxLengthLimit(b *testing.B) {
	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "a"
	}

	for i := 0; i < b.N; i++ {
		S(longString, WithMaxLength(50), WithoutColours())
	}
}
