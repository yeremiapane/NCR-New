package ai

import "fmt"

// SystemPrompt defines the AI's role and output format for NCR analysis
const SystemPrompt = `Anda adalah analis data NCR (Laporan Ketidaksesuaian) yang ahli di perusahaan manufaktur. Tugas Anda adalah menganalisis data dashboard NCR dan memberikan wawasan yang dapat ditindaklanjuti.

Anda akan menerima statistik NCR yang telah digabungkan dan harus menghasilkan wawasan dalam kategori berikut:
1. TREND - Pola seiring waktu, masalah yang meningkat/menurun  
2. PROBLEM - Masalah kritis yang memerlukan perhatian, masalah yang berulang  
3. STATISTIC - Temuan statistik yang menonjol, perbandingan, rasio  
4. RECOMMENDATION - Saran yang dapat ditindaklanjuti untuk mengurangi NCR

RESPONSE FORMAT (array JSON yang ketat):
[
  {
    "type": "TREND|PROBLEM|STATISTIC|RECOMMENDATION",
    "title": "Judul Singkat (max 60 karakter)",
    "description": "Penjelasan wawasan yang detail (2-3 kalimat)",
    "severity": "info|warning|critical"
  }
]

GUIDE KESEITIAN:
- info: Pengamatan umum, temuan netral
- warning: Masalah yang memerlukan perhatian, kekhawatiran moderat
- critical: Masalah mendesak, tren negatif yang signifikan, tindakan segera diperlukan

Aturan:
- Hasilkan 4-8 wawasan total, minimal satu per kategori
- Gunakan angka dan persentase secara spesifik jika tersedia
- Fokus pada wawasan yang dapat ditindaklanjuti dan praktis
- Gunakan bahasa profesional namun jelas
- Hasilkan HANYA array JSON yang valid, tidak ada teks lain`

// BuildAnalysisPrompt creates the user prompt with NCR data context
func BuildAnalysisPrompt(ctx AnalysisContext) string {
	prompt := fmt.Sprintf(`Analyze the following NCR Dashboard data and provide insights:

## SUMMARY STATISTICS
- Total NCRs: %d
- Running (In Progress): %d
- Completed: %d
- Terminated (Rejected): %d
- Approved: %d
- Rejected: %d
- TO (Material Loss): %d
- Non-TO (Rework/Time Loss): %d

## TO vs Non-TO Analysis
Material Loss Rate: %.1f%% dari NCRs yang menyebabkan kerugian material (TO)
Rework Rate: %.1f%% adalah masalah rework/time loss (Non-TO)
`,
		ctx.TotalNCR,
		ctx.RunningCount,
		ctx.CompletedCount,
		ctx.TerminatedCount,
		ctx.ApprovedCount,
		ctx.RejectedCount,
		ctx.TOCount,
		ctx.NonTOCount,
		calculatePercentage(ctx.TOCount, ctx.TotalNCR),
		calculatePercentage(ctx.NonTOCount, ctx.TotalNCR),
	)

	// Add trend data
	if len(ctx.TrendData) > 0 {
		prompt += "\n## MONTHLY/DAILY TREND\n"
		for _, t := range ctx.TrendData {
			prompt += fmt.Sprintf("- %s: %d NCRs\n", t.Period, t.Count)
		}
	}

	// Add top categories
	if len(ctx.TopCategories) > 0 {
		prompt += "\n## TOP NCR CATEGORIES\n"
		for i, cat := range ctx.TopCategories {
			prompt += fmt.Sprintf("%d. %s: %d NCRs (%.1f%%)\n",
				i+1, cat.Name, cat.Count, calculatePercentage(cat.Count, ctx.TotalNCR))
		}
	}

	// Add top brands
	if len(ctx.TopBrands) > 0 {
		prompt += "\n## TOP AFFECTED BRANDS\n"
		for i, brand := range ctx.TopBrands {
			prompt += fmt.Sprintf("%d. %s: %d NCRs\n", i+1, brand.Name, brand.Count)
		}
	}

	// Add top departments
	if len(ctx.TopDepartments) > 0 {
		prompt += "\n## TOP REPORTING DEPARTMENTS\n"
		for i, dept := range ctx.TopDepartments {
			prompt += fmt.Sprintf("%d. %s: %d NCRs\n", i+1, dept.Name, dept.Count)
		}
	}

	// Add top problems
	if len(ctx.TopProblems) > 0 {
		prompt += "\n## TOP RECURRING PROBLEMS\n"
		for i, prob := range ctx.TopProblems {
			prompt += fmt.Sprintf("%d. [%s] %s\n", i+1, prob.Category, prob.Description)
		}
	}

	// Add filter context if present
	if ctx.DateRange != "" {
		prompt += fmt.Sprintf("\n## ANALYSIS PERIOD: %s\n", ctx.DateRange)
	}
	if ctx.Filters != "" {
		prompt += fmt.Sprintf("## ACTIVE FILTERS: %s\n", ctx.Filters)
	}

	prompt += `
Berdasarkan data ini, hasilkan wawasan yang mencakup:
1. Pola tren (meningkat/menurun, anomali)
2. Masalah kritis yang memerlukan perhatian
3. Statistik dan perbandingan yang menonjol
4. Rekomendasi spesifik untuk mengurangi kejadian NCR

Ingat: Output HANYA array JSON yang valid, tidak ada teks lain.`

	return prompt
}

func calculatePercentage(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
