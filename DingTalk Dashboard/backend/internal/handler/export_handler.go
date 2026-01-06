package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	"dingtalk-dashboard/internal/domain/approval"
)

// ExportHandler handles Excel export endpoints
type ExportHandler struct {
	service *approval.Service
}

// NewExportHandler creates a new export handler
func NewExportHandler(service *approval.Service) *ExportHandler {
	return &ExportHandler{service: service}
}

// ExportApprovals handles GET /api/v1/approvals/export
func (h *ExportHandler) ExportApprovals(c *fiber.Ctx) error {
	// Parse filter parameters (same as list approvals)
	params := approval.ListParams{
		Search:          c.Query("search"),
		BusinessID:      c.Query("business_id"),
		Status:          c.Query("status"),
		Department:      c.Query("department"),
		DitujukanKepada: c.Query("ditujukan_kepada"),
		DilaporkanOleh:  c.Query("dilaporkan_oleh"),
		Kategori:        c.Query("kategori"),
		Page:            1,
		PageSize:        10000, // Export all matching records
	}

	// Parse dates
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			params.EndDate = &t
		}
	}

	// Get all matching approvals
	approvals, _, err := h.service.ListApprovals(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch data for export",
			"error":   err.Error(),
		})
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "NCR Data"
	f.SetSheetName("Sheet1", sheetName)

	// Define styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  11,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F46E5"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#3730A3", Style: 1},
			{Type: "right", Color: "#3730A3", Style: 1},
			{Type: "top", Color: "#3730A3", Style: 1},
			{Type: "bottom", Color: "#3730A3", Style: 1},
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	altDataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F9FAFB"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	statusRunningStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#B45309"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FEF3C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	statusApprovedStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#047857"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D1FAE5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	statusRejectedStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#DC2626"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FEE2E2"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	linkStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Color: "#2563EB", Underline: "single"},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})

	// Set column widths for all columns
	colWidths := map[string]float64{
		"A": 15, // Business ID
		"B": 12, // Tanggal
		"C": 12, // Status
		"D": 10, // Result
		"E": 20, // Department
		"F": 15, // Originator Name
		"G": 15, // Kategori
		"H": 25, // Nama Project
		"I": 15, // Nomor FPPP
		"J": 15, // Nomor PO
		"K": 25, // Nama Item Product
		"L": 20, // Ditujukan Kepada
		"M": 20, // Dilaporkan Oleh
		"N": 10, // TO/Tidak TO
		"O": 15, // Urgent Butuh Kapan
		"P": 40, // Deskripsi Masalah
		"Q": 30, // Catatan Tambahan
		"R": 30, // Detail Material
		"S": 30, // Analisis Penyebab
		"T": 20, // Nama Melakukan Masalah
		"U": 30, // Tindakan Perbaikan
		"V": 30, // Tindakan Pencegahan
		"W": 40, // Remark Comment
		"X": 50, // Attachment URLs
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// Set header row - ALL fields
	headers := []string{
		"Business ID", "Tanggal", "Status", "Result", "Department",
		"Originator Name", "Kategori", "Nama Project", "Nomor FPPP", "Nomor PO",
		"Nama Item/Product", "Ditujukan Kepada", "Dilaporkan Oleh", "TO/Tidak TO",
		"Urgent Butuh Kapan", "Deskripsi Masalah", "Catatan Tambahan", "Detail Material",
		"Analisis Penyebab", "Nama Melakukan Masalah", "Tindakan Perbaikan",
		"Tindakan Pencegahan", "Remark Comment", "Attachments/Photos",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 1, 30)

	// Add data rows
	for rowIdx, appr := range approvals {
		row := rowIdx + 2

		// Select alternating row style
		rowStyle := dataStyle
		if rowIdx%2 == 1 {
			rowStyle = altDataStyle
		}

		// A: Business ID
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), appr.BusinessID)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), rowStyle)

		// B: Tanggal
		tanggal := ""
		if appr.Tanggal != nil {
			tanggal = appr.Tanggal.Format("02-Jan-2006")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), tanggal)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), rowStyle)

		// C: Status with conditional formatting
		statusStyle := rowStyle
		statusText := appr.Status
		if appr.Result == "agree" {
			statusText = "Approved"
			statusStyle = statusApprovedStyle
		} else if appr.Result == "refuse" {
			statusText = "Rejected"
			statusStyle = statusRejectedStyle
		} else if appr.Status == "RUNNING" {
			statusText = "Running"
			statusStyle = statusRunningStyle
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), statusText)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), statusStyle)

		// D: Result
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), appr.Result)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), rowStyle)

		// E: Department
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), appr.OriginatorDeptName)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), rowStyle)

		// F: Originator Name
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), appr.OriginatorName)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), rowStyle)

		// G: Kategori
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), appr.Kategori)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), rowStyle)

		// H: Nama Project
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), appr.NamaProject)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), rowStyle)

		// I: Nomor FPPP
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), appr.NomorFPPP)
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), rowStyle)

		// J: Nomor Production Order
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), appr.NomorProductionOrder)
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), rowStyle)

		// K: Nama Item Product
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), appr.NamaItemProduct)
		f.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), rowStyle)

		// L: Ditujukan Kepada
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), appr.DitujukanKepada)
		f.SetCellStyle(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), rowStyle)

		// M: Dilaporkan Oleh
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), appr.DilaporkanOleh)
		f.SetCellStyle(sheetName, fmt.Sprintf("M%d", row), fmt.Sprintf("M%d", row), rowStyle)

		// N: TO/Tidak TO
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), appr.ToTidakTo)
		f.SetCellStyle(sheetName, fmt.Sprintf("N%d", row), fmt.Sprintf("N%d", row), rowStyle)

		// O: Urgent Butuh Kapan
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), appr.UrgentButuhKapan)
		f.SetCellStyle(sheetName, fmt.Sprintf("O%d", row), fmt.Sprintf("O%d", row), rowStyle)

		// P: Deskripsi Masalah
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), appr.DeskripsiMasalah)
		f.SetCellStyle(sheetName, fmt.Sprintf("P%d", row), fmt.Sprintf("P%d", row), rowStyle)

		// Q: Catatan Tambahan
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), appr.CatatanTambahan)
		f.SetCellStyle(sheetName, fmt.Sprintf("Q%d", row), fmt.Sprintf("Q%d", row), rowStyle)

		// R: Detail Material Yang Dibutuhkan
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), appr.DetailMaterialYangDibutuhkan)
		f.SetCellStyle(sheetName, fmt.Sprintf("R%d", row), fmt.Sprintf("R%d", row), rowStyle)

		// S: Analisis Penyebab Masalah
		f.SetCellValue(sheetName, fmt.Sprintf("S%d", row), appr.AnalisisPenyebabMasalah)
		f.SetCellStyle(sheetName, fmt.Sprintf("S%d", row), fmt.Sprintf("S%d", row), rowStyle)

		// T: Nama Yang Melakukan Masalah
		f.SetCellValue(sheetName, fmt.Sprintf("T%d", row), appr.NamaYangMelakukanMasalah)
		f.SetCellStyle(sheetName, fmt.Sprintf("T%d", row), fmt.Sprintf("T%d", row), rowStyle)

		// U: Tindakan Perbaikan
		f.SetCellValue(sheetName, fmt.Sprintf("U%d", row), appr.TindakanPerbaikan)
		f.SetCellStyle(sheetName, fmt.Sprintf("U%d", row), fmt.Sprintf("U%d", row), rowStyle)

		// V: Tindakan Pencegahan
		f.SetCellValue(sheetName, fmt.Sprintf("V%d", row), appr.TindakanPencegahan)
		f.SetCellStyle(sheetName, fmt.Sprintf("V%d", row), fmt.Sprintf("V%d", row), rowStyle)

		// W: Remark Comment
		f.SetCellValue(sheetName, fmt.Sprintf("W%d", row), appr.RemarkComment)
		f.SetCellStyle(sheetName, fmt.Sprintf("W%d", row), fmt.Sprintf("W%d", row), rowStyle)

		// X: Attachments/Photos - combine all URLs
		var attachmentURLs []string
		for _, att := range appr.Attachments {
			if att.FileURL != "" {
				attachmentURLs = append(attachmentURLs, att.FileURL)
			}
		}
		attachmentText := strings.Join(attachmentURLs, "\n")
		f.SetCellValue(sheetName, fmt.Sprintf("X%d", row), attachmentText)
		if len(attachmentURLs) > 0 {
			f.SetCellStyle(sheetName, fmt.Sprintf("X%d", row), fmt.Sprintf("X%d", row), linkStyle)
		} else {
			f.SetCellStyle(sheetName, fmt.Sprintf("X%d", row), fmt.Sprintf("X%d", row), rowStyle)
		}

		// Set row height
		f.SetRowHeight(sheetName, row, 25)
	}

	// Freeze header row
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// Generate filename with date
	filename := fmt.Sprintf("NCR_Export_%s.xlsx", time.Now().Format("2006-01-02_150405"))

	// Write to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to generate Excel file",
			"error":   err.Error(),
		})
	}

	// Set response headers
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return c.Send(buffer.Bytes())
}
