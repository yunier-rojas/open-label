package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/go-pdf/fpdf"
	"github.com/yougg/go-qrcode"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

func signPayload(payload string, privateKey []byte, passphrase []byte) ([]byte, error) {
	hash := sha256.Sum256([]byte(payload))
	hashedPayload := fmt.Sprintf("%x", hash)

	armored, err := helper.SignCleartextMessageArmored(string(privateKey), passphrase, hashedPayload)
	if err != nil {
		return []byte(""), err
	}

	clearTextMessage, err := crypto.NewClearTextMessageFromArmored(armored)
	if err != nil {
		return []byte(""), err
	}

	return []byte(base64.StdEncoding.EncodeToString(clearTextMessage.GetBinarySignature())), nil
}

func appendSignature(payload string, signature []byte) string {
	// Convert payload to string and trim any trailing newlines
	yamlStr := strings.TrimSpace(payload)

	// Append signature as new YAML field
	yamlStr += fmt.Sprintf("\nx_sig: %s\n", signature)

	return yamlStr
}

func generateQRCode(payload string, filePath string) error {
	return qrcode.WriteFile(payload, qrcode.Low, 512, filePath, 0)
}

func main() {
	// Read and parse YAML input
	data, err := os.ReadFile("claims-public.yaml")
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}

	privateKey, err := os.ReadFile("private.pgp") // Encrypted private key
	if err != nil {
		log.Fatalf("Unable to read private key file: %v", err)
	}

	passphrase := []byte("qwert1234") // Private key passphrase

	// Sign the payload
	payload := strings.TrimSpace(string(data))
	signature, err := signPayload(payload, privateKey, passphrase)
	if err != nil {
		log.Fatalf("Failed to sign payload: %v", err)
	}

	finalPayload := appendSignature(payload, signature)

	// Write QR Code
	err = generateQRCode(finalPayload, "qr.png")
	if err != nil {
		log.Fatalf("Failed to create QR code: %v", err)
	}

	// Write payload to file
	err = os.WriteFile("payload.yaml", []byte(finalPayload), 0644)
	if err != nil {
		log.Fatalf("Failed to write payload file: %v", err)
	}

	claims := make(map[string]string)
	err = yaml.Unmarshal([]byte(payload), &claims)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	// Generate PDF
	pdf := fpdf.New("P", "mm", "A6", "")
	pdf.AddPage()

	// Add QR code to PDF
	pdf.Image("qr.png", 0, 0, 105, 105, false, "", 0, "")

	// Add recipient details to PDF
	pdf.SetFont("Arial", "B", 15)
	pdf.Text(5, 110, "Recipient")
	pdf.SetFont("Arial", "", 10)
	recipientKeys := []string{"r_person", "r_address"}
	x, y := 5, 115
	for idx, key := range recipientKeys {
		value, ok := claims[key]
		if !ok {
			log.Fatalf("Missing key: %v", key)
		}
		pdf.Text(float64(x), float64(y+idx*4), value)
	}

	// Add sender details to PDF
	pdf.SetFont("Arial", "B", 8)
	pdf.Text(5, 130, "Sender")
	pdf.SetFont("Arial", "", 5)
	senderKeys := []string{"s_name", "s_address"}
	x, y = 5, 134
	for idx, key := range senderKeys {
		value, ok := claims[key]
		if !ok {
			log.Fatalf("Missing key: %v", key)
		}
		pdf.Text(float64(x), float64(y+idx*2), value)
	}

	err = pdf.OutputFileAndClose("label.pdf")
	if err != nil {
		log.Fatalf("Failed to output PDF file: %v", err)
	}
}
