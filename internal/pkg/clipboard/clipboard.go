package clipboard

import (
	atottoclip "github.com/atotto/clipboard"
)

// Write copies text to the system clipboard using OS-native mechanisms.
func Write(text string) error {
	return atottoclip.WriteAll(text)
}
