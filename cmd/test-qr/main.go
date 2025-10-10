package main

import (
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"
)

func main() {
	// Test 1: Simple URL (should work)
	fmt.Println("Test 1: Generating QR from simple URL...")
	err := qrcode.WriteFile("2@/DnbWHew5R4K9cSWpnvrdJJEKwvK8SDIqFzBwMnsUbS/PopqGtWx7W25Hs+shxum0xOxFeD3QXMg/o/3uD79qq8i+i8Vh1H+xi8=,p7OL1IJqfyjGBk5qIS+lbt9WgF8CpePwF8TlYr+96EE=,zOidCE/VM2cZsAkONJ8cz2vrBAf25QhZHu4tbTWF9m0=,", qrcode.Medium, 256, "test-url.png")
	if err != nil {
		fmt.Printf("❌ Failed to generate URL QR: %v\n", err)
	} else {
		fmt.Println("✅ URL QR generated: test-url.png")
	}

	// Test 2: WhatsApp format QR (typical format from whatsmeow)
	// Format: "2@..." followed by base64-like string
	fmt.Println("\nTest 2: Generating QR from WhatsApp-like string...")
	whatsappQR := "2@ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=="
	err = qrcode.WriteFile(whatsappQR, qrcode.Medium, 256, "test-whatsapp.png")
	if err != nil {
		fmt.Printf("❌ Failed to generate WhatsApp QR: %v\n", err)
	} else {
		fmt.Println("✅ WhatsApp QR generated: test-whatsapp.png")
	}

	// Test 3: Generate in memory and check size
	fmt.Println("\nTest 3: Generating QR in memory...")
	qrBytes, err := qrcode.Encode(whatsappQR, qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("❌ Failed to encode: %v\n", err)
	} else {
		fmt.Printf("✅ QR encoded successfully, size: %d bytes\n", len(qrBytes))
	}

	// Test 4: Empty string (should fail)
	fmt.Println("\nTest 4: Generating QR from empty string...")
	qrBytes, err = qrcode.Encode("", qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("✅ Expected failure for empty string: %v\n", err)
	} else {
		fmt.Printf("⚠️  Unexpected success with empty string, size: %d bytes\n", len(qrBytes))
	}

	// Test 5: Very long string (WhatsApp QR codes are typically long)
	fmt.Println("\nTest 5: Generating QR from long string...")
	longString := "2@" + string(make([]byte, 500)) // Simulate long QR
	for i := range longString[2:] {
		longString = longString[:2+i] + "A" + longString[3+i:]
	}
	qrBytes, err = qrcode.Encode(longString, qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("❌ Failed to encode long string: %v\n", err)
	} else {
		fmt.Printf("✅ Long string encoded successfully, size: %d bytes\n", len(qrBytes))
		err = os.WriteFile("test-long.png", qrBytes, 0644)
		if err != nil {
			fmt.Printf("❌ Failed to write file: %v\n", err)
		} else {
			fmt.Println("✅ Long QR written to: test-long.png")
		}
	}

	fmt.Println("\n=================================")
	fmt.Println("Test Summary:")
	fmt.Println("Check generated files:")
	fmt.Println("- test-url.png")
	fmt.Println("- test-whatsapp.png")
	fmt.Println("- test-long.png")
	fmt.Println("\nIf files exist and can be scanned, library is working correctly!")
}
