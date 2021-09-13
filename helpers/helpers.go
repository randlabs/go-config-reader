package helpers

import (
	"fmt"
)

// -----------------------------------------------------------------------------

func LoadError(err error) error {
	return fmt.Errorf("unable to load configuration [%v]", err)
}

func RemoveComments(data []byte) {
	state := 0

	dataLength := len(data)

	for index := 0; index < dataLength; index++ {
		ch := data[index]

		switch state {
		case 0: // Outside quotes and comments
			if ch == '"' {
				state = 1 // Start of quoted value
			} else if ch == '/' {
				if index + 1 < dataLength {
					switch data[index + 1] {
					case '/':
						state = 2 // Start of single line comment
						data[index] = ' '
						index += 1
						data[index] = ' '

					case '*':
						state = 3 // Start of multiple line comment
						data[index] = ' '
						index += 1
						data[index] = ' '
					}
				}
			}

		case 1: // Inside quoted value
			if ch == '\\' {
				if index + 1 < dataLength {
					index += 1 // Escaped character
				}
			} else if ch == '"' {
				state = 0 // End of quoted value
			}

		case 2: // Single line comment
			if ch == '\n' {
				state = 0 // End of single line comment
			} else {
				data[index] = ' ' // Remove comment
			}

		case 3: // Multiple line comment
			if ch == '*' && index + 1 < dataLength && data[index + 1] == '/' {
				state = 0 // End of single line comment
				index += 1
			} else {
				data[index] = ' ' // Remove comment
			}
		}
	}
}
