package spinner

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

func Spinner(ch chan (bool)) {
	// Define the spinner characters
	spinnerChars := []rune{'|', '/', '-', '\\', '|', '/', '-', '\\'}

	// Run the spinner
	fmt.Print("\n")
	for {
		for i, r := range spinnerChars {
			select {
			case <-ch:
				return
			default:
				fmt.Printf("\rWait %c %s%s", r, strings.Repeat(".", i+1), strings.Repeat(" ", slices.Max([]int{0, len(spinnerChars) - i - 1})))
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
}
