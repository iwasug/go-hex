package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"
)

func main() {
	qrString := flag.String("qr", "", "QR code string to generate (required)")
	output := flag.String("output", "manual-qr.png", "Output PNG file name")
	size := flag.Int("size", 256, "QR code size in pixels")
	
	flag.Parse()

	if *qrString == "" {
		fmt.Println("❌ Error: -qr parameter is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run cmd/qr-manual/main.go -qr \"YOUR_QR_STRING\" -output qr.png -size 256")
		fmt.Println("\nExample:")
		fmt.Println("  go run cmd/qr-manual/main.go -qr \"2@ABC123...\" -output whatsapp-qr.png")
		fmt.Println("\nGet QR string from API:")
		fmt.Println("  curl -H \"Authorization: Bearer TOKEN\" \\")
		fmt.Println("       \"http://localhost:8080/api/v1/whatsapp/sessions/{id}/qr?format=json\"")
		os.Exit(1)
	}

	fmt.Printf("Generating QR code...\n")
	fmt.Printf("- Input length: %d characters\n", len(*qrString))
	fmt.Printf("- Output file: %s\n", *output)
	fmt.Printf("- Size: %dx%d pixels\n", *size, *size)
	fmt.Println()

	// Generate QR code
	err := qrcode.WriteFile(*qrString, qrcode.Medium, *size, *output)
	if err != nil {
		fmt.Printf("❌ Failed to generate QR code: %v\n", err)
		os.Exit(1)
	}

	// Check file size
	fileInfo, err := os.Stat(*output)
	if err != nil {
		fmt.Printf("⚠️  QR generated but cannot verify file: %v\n", err)
		os.Exit(0)
	}

	fmt.Printf("✅ QR code generated successfully!\n")
	fmt.Printf("   File: %s\n", *output)
	fmt.Printf("   Size: %d bytes\n", fileInfo.Size())
	fmt.Println()
	fmt.Println("Scan this QR code with WhatsApp to link your account.")
	fmt.Println()
	fmt.Println("Note: QR codes from WhatsApp expire after 60 seconds (default).")
	fmt.Println("      Generate a new one if expired.")
}
