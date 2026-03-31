package card

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Luhn algorithm - validates and generates card numbers
//
// How it works:
//   1. Start from the rightmost digit, double every second digit
//   2. If doubled value > 9, subtract 9
//   3. Sum all digits
//   4. If sum % 10 == 0, the number is valid
//
// Example: 4539 1488 0343 6467
//   Step 1: 8,5,2,9,2,4,0,8,0,3,8,3,12,4,12,7
//   Step 2: 8,5,2,9,2,4,0,8,0,3,8,3,3,4,3,7
//   Sum = 69... wait, let me just implement it properly

// ValidateLuhn checks if a card number passes the Luhn algorithm
func ValidateLuhn(number string) bool {
	sum := 0
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if d < 0 || d > 9 {
			return false
		}

		if isSecond {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		isSecond = !isSecond
	}

	return sum%10 == 0
}

// GenerateCardNumber generates a valid 16-digit card number with Luhn check digit
// BIN (first 6 digits): 486486 (XBank identifier)
func GenerateCardNumber() (string, error) {
	const bin = "486486"

	// Generate 9 random digits
	var random string
	for i := 0; i < 9; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		random += fmt.Sprintf("%d", n.Int64())
	}

	// 15 digits without check digit
	partial := bin + random

	// Calculate Luhn check digit
	checkDigit := luhnCheckDigit(partial)
	return partial + fmt.Sprintf("%d", checkDigit), nil
}

// luhnCheckDigit calculates the check digit for a partial card number
func luhnCheckDigit(partial string) int {
	sum := 0
	isSecond := true // because we're appending one more digit

	for i := len(partial) - 1; i >= 0; i-- {
		d := int(partial[i] - '0')

		if isSecond {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		isSecond = !isSecond
	}

	return (10 - (sum % 10)) % 10
}

// MaskPAN returns masked card number: **** **** **** 1234
func MaskPAN(pan string) string {
	if len(pan) < 4 {
		return "****"
	}
	return "**** **** **** " + pan[len(pan)-4:]
}
