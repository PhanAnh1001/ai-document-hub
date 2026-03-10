package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// XMLInvoiceParser parses Vietnamese electronic invoice (hóa đơn điện tử) XML files.
// Supports the TCVN e-invoice format with root element <HDon>.
type XMLInvoiceParser struct{}

func NewXMLInvoiceParser() *XMLInvoiceParser {
	return &XMLInvoiceParser{}
}

// hdon is the Go struct mapping to the Vietnamese e-invoice XML schema.
type hdon struct {
	XMLName  xml.Name  `xml:"HDon"`
	TTChung  ttChung   `xml:"TTChung"`
	NMua     partyInfo `xml:"NMua"`
	NBan     partyInfo `xml:"NBan"`
	DSHHDVu  dshhDVu   `xml:"DSHHDVu"`
	TToan    tToan     `xml:"TToan"`
}

type ttChung struct {
	SHDon  string `xml:"SHDon"`  // Invoice number
	NLap   string `xml:"NLap"`   // Issue date
}

type partyInfo struct {
	Ten  string `xml:"Ten"`  // Name
	MST  string `xml:"MST"`  // Tax code
	DChi string `xml:"DChi"` // Address
}

type dshhDVu struct {
	Items []hhDVu `xml:"HHDVu"`
}

type hhDVu struct {
	STT     string  `xml:"STT"`     // Line number
	THHDVu  string  `xml:"THHDVu"`  // Description
	DVTinh  string  `xml:"DVTinh"`  // Unit
	SLuong  float64 `xml:"SLuong"`  // Quantity
	DGia    float64 `xml:"DGia"`    // Unit price
	ThTien  float64 `xml:"ThTien"`  // Line total
}

type tToan struct {
	TgTCThue  string `xml:"TgTCThue"`  // Total before tax
	TgTThue   string `xml:"TgTThue"`   // Tax amount
	TgTTTBSo  string `xml:"TgTTTBSo"`  // Total with tax
}

// ExtractFromFile reads and parses an XML e-invoice file, returning a normalized OCRResult.
func (p *XMLInvoiceParser) ExtractFromFile(_ context.Context, filePath, docType string) (*OCRResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file: %w", err)
	}

	var inv hdon
	if err := xml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse xml invoice: %w", err)
	}

	// Validate parsed result has minimum required fields
	if inv.NBan.Ten == "" && inv.NMua.Ten == "" {
		return nil, fmt.Errorf("invalid xml invoice: missing buyer/seller info")
	}

	totalAmount := parseXMLAmount(inv.TToan.TgTTTBSo)
	taxAmount := parseXMLAmount(inv.TToan.TgTThue)

	lineItems := make([]OCRLineItem, 0, len(inv.DSHHDVu.Items))
	for _, item := range inv.DSHHDVu.Items {
		lineItems = append(lineItems, OCRLineItem{
			Description: item.THHDVu,
			Quantity:    item.SLuong,
			UnitPrice:   item.DGia,
			Amount:      item.ThTien,
		})
	}

	return &OCRResult{
		DocumentType: docType,
		Vendor:       inv.NBan.Ten, // Seller is the vendor on a purchase invoice
		TotalAmount:  totalAmount,
		TaxAmount:    taxAmount,
		Date:         normDate(inv.TTChung.NLap),
		Currency:     "VND",
		LineItems:    lineItems,
		RawData: map[string]interface{}{
			"invoice_no": inv.TTChung.SHDon,
			"buyer":      inv.NMua.Ten,
			"seller":     inv.NBan.Ten,
		},
	}, nil
}

// parseXMLAmount converts amount strings from XML (no commas, just digits) to float64.
func parseXMLAmount(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
