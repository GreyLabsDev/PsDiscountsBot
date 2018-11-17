package stringtool

import (
	"strings"
)

/*ExtractBetween ...
 */
func ExtractBetween(startChars string, endChars string, initialString string) (result string) {
	if !strings.Contains(initialString, startChars) && !strings.Contains(initialString, endChars) {
		return
	} else if !strings.Contains(initialString, startChars) {
		return
	} else if !strings.Contains(initialString, endChars) {
		return
	}
	startIndex := strings.Index(initialString, startChars) + len(startChars)
	if startIndex == -1 {
		return
	}
	endIndex := strings.Index(initialString, endChars)
	if endIndex >= len(initialString) {
		return
	}

	return initialString[startIndex:endIndex]
}
