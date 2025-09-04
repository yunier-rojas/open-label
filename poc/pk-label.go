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
	yamlStr := strings.TrimSpace(payload)
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

	privateKey, err := os.ReadFile("private.pgp")
	if err != nil {
		log.Fatalf("Unable to read private key file: %v", err)
	}

	passphrase := []byte("qwert1234")

	// Sign the payload
	payload := strings.TrimSpace(string(data))
	signature, err := signPayload(payload, privateKey, passphrase)
	if err != nil {
		log.Fatalf("Failed to sign payload: %v", err)
	}

	finalPayload := appendSignature(payload, signature)

	// Write QR code
	err = generateQRCode(finalPayload, "qr.png")
	if err != nil {
		log.Fatalf("Failed to create QR code: %v", err)
	}

	// Write payload to file
	err = os.WriteFile("payload.yaml", []byte(finalPayload), 0644)
	if err != nil {
		log.Fatalf("Failed to write payload file: %v", err)
	}

	// Parse claims
	claims := make(map[string]string)
	err = yaml.Unmarshal([]byte(payload), &claims)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	// === Generate Landscape PDF Label ===
	pdf := fpdf.New("L", "mm", "A6", "")
	pdf.AddPage()
	width, height := pdf.GetPageSize()

	// QR code on left
	pdf.Image("qr.png", 2, 2, 100, 100, false, "", 0, "")

	// Recipient details on right top, but transposed
	pdf.TransformBegin()
	pdf.TransformRotate(-90, width/2, width/2)

	pdf.SetFont("Arial", "B", 16)
	pdf.Text(6, 12, "Recipient")
	pdf.SetFont("Arial", "", 12)
	recipientKeys := []string{"r_person", "r_address"}
	y := 20
	for _, key := range recipientKeys {
		value := claims[key]
		if value == "" {
			value = "[missing]"
		}
		pdf.Text(6, float64(y), value)
		y += 6
	}
	// Draw a rectangle for the recipient section
	pdf.Rect(4, 4, height-10, 30, "D")

	// Sender details
	pdf.SetFont("Arial", "B", 10)
	pdf.Text(6, 40, "Sender")
	pdf.SetFont("Arial", "", 8)
	sender := claims["s_name"]
	pdf.Text(6, 45, sender)
	senderAddress := claims["s_address"]
	pdf.Text(35, 45, senderAddress)

	pdf.Line(30, 37, 30, 47)

	pdf.TransformEnd()

	// Output PDF
	err = pdf.OutputFileAndClose("label.pdf")
	if err != nil {
		log.Fatalf("Failed to output PDF file: %v", err)
	}
}
