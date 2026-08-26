package clipboard

import (
	"testing"
)

func TestClipboardWrite(t *testing.T) {
	// Write should not panic
	_ = Write("monobox clipboard test")
}
