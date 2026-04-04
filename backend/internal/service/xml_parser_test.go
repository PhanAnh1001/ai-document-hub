package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const sampleXMLInvoice = `<?xml version="1.0" encoding="UTF-8"?>
<HDon>
  <TTChung>
    <SHDon>0001</SHDon>
    <NLap>28/03/2026</NLap>
    <KHMSHDon>01GTKT0/001</KHMSHDon>
  </TTChung>
  <NMua>
    <Ten>Công ty TNHH ABC</Ten>
    <MST>0123456789</MST>
    <DChi>123 Đường ABC, Q1, TP.HCM</DChi>
  </NMua>
  <NBan>
    <Ten>Công ty CP XYZ</Ten>
    <MST>9876543210</MST>
    <DChi>456 Đường XYZ, Q3, TP.HCM</DChi>
  </NBan>
  <DSHHDVu>
    <HHDVu>
      <STT>1</STT>
      <THHDVu>Hàng hóa nhập kho</THHDVu>
      <DVTinh>Cái</DVTinh>
      <SLuong>10</SLuong>
      <DGia>1000000</DGia>
      <ThTien>10000000</ThTien>
    </HHDVu>
  </DSHHDVu>
  <TToan>
    <TgTCThue>10000000</TgTCThue>
    <TgTThue>1000000</TgTThue>
    <TgTTTBSo>11000000</TgTTTBSo>
  </TToan>
</HDon>`

func writeTempXML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "invoice.xml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp xml: %v", err)
	}
	return path
}

func TestXMLInvoiceParser_ExtractFromFile_ParsesVendor(t *testing.T) {
	path := writeTempXML(t, sampleXMLInvoice)
	parser := NewXMLInvoiceParser()

	result, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Vendor != "Công ty CP XYZ" {
		t.Errorf("expected vendor 'Công ty CP XYZ', got %q", result.Vendor)
	}
}

func TestXMLInvoiceParser_ExtractFromFile_ParsesTotalAmount(t *testing.T) {
	path := writeTempXML(t, sampleXMLInvoice)
	parser := NewXMLInvoiceParser()

	result, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TgTTTBSo = 11000000 (total with tax)
	if result.TotalAmount != 11000000 {
		t.Errorf("expected total 11000000, got %v", result.TotalAmount)
	}
}

func TestXMLInvoiceParser_ExtractFromFile_ParsesTaxAmount(t *testing.T) {
	path := writeTempXML(t, sampleXMLInvoice)
	parser := NewXMLInvoiceParser()

	result, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TaxAmount != 1000000 {
		t.Errorf("expected tax 1000000, got %v", result.TaxAmount)
	}
}

func TestXMLInvoiceParser_ExtractFromFile_ParsesDate(t *testing.T) {
	path := writeTempXML(t, sampleXMLInvoice)
	parser := NewXMLInvoiceParser()

	result, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Date != "2026-03-28" {
		t.Errorf("expected date '2026-03-28', got %q", result.Date)
	}
}

func TestXMLInvoiceParser_ExtractFromFile_ParsesLineItems(t *testing.T) {
	path := writeTempXML(t, sampleXMLInvoice)
	parser := NewXMLInvoiceParser()

	result, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(result.LineItems))
	}
	item := result.LineItems[0]
	if item.Description != "Hàng hóa nhập kho" {
		t.Errorf("expected description 'Hàng hóa nhập kho', got %q", item.Description)
	}
	if item.Amount != 10000000 {
		t.Errorf("expected line item amount 10000000, got %v", item.Amount)
	}
}

func TestXMLInvoiceParser_ExtractFromFile_FileNotFound(t *testing.T) {
	parser := NewXMLInvoiceParser()
	_, err := parser.ExtractFromFile(context.Background(), "/nonexistent/file.xml", "invoice")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestXMLInvoiceParser_ExtractFromFile_InvalidXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.xml")
	os.WriteFile(path, []byte("not valid xml <<<"), 0644)

	parser := NewXMLInvoiceParser()
	_, err := parser.ExtractFromFile(context.Background(), path, "invoice")
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}
