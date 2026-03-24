package loader

import (
	"fmt"
	"testing"
)

func TestLoader(t *testing.T) {
	proj, err := Load([]byte(Sample))
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	fmt.Printf("Loaded project: %v\n", proj)
}
