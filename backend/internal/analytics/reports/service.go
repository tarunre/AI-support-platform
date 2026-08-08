package reports

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Service manages report generation, file storage, and history.
type Service struct {
	db            *gorm.DB
	storagePath   string
	retentionDays int
}

// NewService creates a ReportService.
func NewService(db *gorm.DB, storagePath string, retentionDays int) *Service {
	return &Service{db: db, storagePath: storagePath, retentionDays: retentionDays}
}

// ReportData is a generic container passed to format-specific generators.
type ReportData struct {
	Title   string
	Headers []string
	Rows    [][]string
	Meta    map[string]string
}

// Schedule creates a PENDING report record and launches generation in a goroutine.
func (s *Service) Schedule(req *dto.GenerateReportRequest, tenantID uuid.UUID, userID uint, svc interface{ CollectData(string, dto.DateFilter) (*ReportData, error) }) (*models.Report, error) {
	filtersJSON, _ := json.Marshal(req)
	report := &models.Report{
		Name:        req.Name,
		Type:        req.ReportType,
		Format:      req.Format,
		Status:      "PENDING",
		Parameters:  string(filtersJSON),
		GeneratedBy: userID,
		TenantID:    tenantID,
	}
	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}

	filter := filterFromReq(req)
	go s.generate(report.ID, filter, svc)

	return report, nil
}

func (s *Service) generate(reportID uint, filter dto.DateFilter, svc interface{ CollectData(string, dto.DateFilter) (*ReportData, error) }) {
	var report models.Report
	if err := s.db.First(&report, reportID).Error; err != nil {
		utils.Logger.WithError(err).Error("Reports: report record not found")
		return
	}

	data, err := svc.CollectData(report.Type, filter)
	if err != nil {
		s.markFailed(&report, err.Error())
		return
	}

	var fileBytes []byte
	var ext string
	switch report.Format {
	case "CSV":
		fileBytes, err = generateCSV(data)
		ext = ".csv"
	case "EXCEL":
		fileBytes, err = generateExcel(data)
		ext = ".xlsx"
	default:
		fileBytes, err = generateHTML(data)
		ext = ".html"
	}
	if err != nil {
		s.markFailed(&report, err.Error())
		return
	}

	if err := os.MkdirAll(s.storagePath, 0750); err != nil {
		s.markFailed(&report, fmt.Sprintf("mkdir: %v", err))
		return
	}

	filename := fmt.Sprintf("report_%d_%d%s", reportID, time.Now().Unix(), ext)
	filePath := filepath.Join(s.storagePath, filename)

	if err := os.WriteFile(filePath, fileBytes, 0600); err != nil {
		s.markFailed(&report, fmt.Sprintf("write: %v", err))
		return
	}

	report.Status = "COMPLETED"
	report.FilePath = filePath
	report.FileSize = int64(len(fileBytes))
	// CompletedAt removed
	s.db.Save(&report)
}

func (s *Service) markFailed(report *models.Report, reason string) {
        report.Status = "FAILED"
        report.ErrorMsg = reason
}

// GetReport returns a single report by ID.
func (s *Service) GetReport(tenantID uuid.UUID, id uint) (*models.Report, error) {
	var report models.Report
	err := s.db.Preload("Generator").Where("id = ? AND tenant_id = ?", id, tenantID).First(&report).Error
	return &report, err
}

// ListReports returns all reports scoped to a tenant, optionally filtered to a single user.
func (s *Service) ListReports(tenantID uuid.UUID, generatedBy *uint) ([]models.Report, error) {
	var reports []models.Report
	q := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC")
	if generatedBy != nil {
		q = q.Where("generated_by = ?", *generatedBy)
	}
	return reports, q.Find(&reports).Error
}

// DownloadReport returns the file bytes and MIME type for a completed report.
func (s *Service) DownloadReport(tenantID uuid.UUID, id uint) ([]byte, string, string, error) {
	var report models.Report
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&report).Error; err != nil {
		return nil, "", "", err
	}
	if report.Status != "COMPLETED" {
		return nil, "", "", fmt.Errorf("report not yet completed (status: %s)", report.Status)
	}
	data, err := os.ReadFile(report.FilePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("file not found: %w", err)
	}

	var mime, filename string
	switch report.Format {
	case "CSV":
		mime = "text/csv"
		filename = report.Name + ".csv"
	case "EXCEL":
		mime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = report.Name + ".xlsx"
	default:
		mime = "text/html; charset=utf-8"
		filename = report.Name + ".html"
	}
	return data, mime, filename, nil
}

// CleanupOldReports deletes files and records older than retentionDays.
func (s *Service) CleanupOldReports() {
	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	var reports []models.Report
	if err := s.db.Where("created_at < ?", cutoff).Find(&reports).Error; err != nil {
		return
	}
	for _, report := range reports {
		if report.FilePath != "" {
			_ = os.Remove(report.FilePath)
		}
		s.db.Delete(&models.Report{}, report.ID)
	}
}

// ─── Format generators ───────────────────────────────────────────────────────

func generateCSV(data *ReportData) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(data.Headers); err != nil {
		return nil, err
	}
	for _, row := range data.Rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func generateExcel(data *ReportData) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Report"
	_ = f.SetSheetName("Sheet1", sheet)

	// Style header row
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E79"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
		"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T"}
	for i, h := range data.Headers {
		if i >= len(cols) {
			break
		}
		cell := cols[i] + "1"
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
		_ = f.SetColWidth(sheet, cols[i], cols[i], 18)
	}

	for rowIdx, row := range data.Rows {
		for colIdx, val := range row {
			if colIdx >= len(cols) {
				break
			}
			cell := fmt.Sprintf("%s%d", cols[colIdx], rowIdx+2)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}

	// Add title above table
	_ = f.SetCellValue(sheet, "A1", data.Title) // overwrite first cell with title handled via sheet rename

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var htmlTmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>{{.Title}}</title>
<style>
  body { font-family: Arial, sans-serif; margin: 32px; color: #111; }
  h1 { font-size: 22px; color: #1F4E79; border-bottom: 2px solid #1F4E79; padding-bottom: 8px; }
  .meta { font-size: 12px; color: #555; margin-bottom: 20px; }
  table { border-collapse: collapse; width: 100%; font-size: 13px; }
  th { background: #1F4E79; color: white; padding: 8px 12px; text-align: left; }
  td { padding: 7px 12px; border-bottom: 1px solid #ddd; }
  tr:nth-child(even) { background: #f2f6fc; }
  @media print { body { margin: 0; } }
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<div class="meta">
{{range $k,$v := .Meta}}<span><strong>{{$k}}:</strong> {{$v}}</span> &nbsp; {{end}}
</div>
<table>
  <thead><tr>{{range .Headers}}<th>{{.}}</th>{{end}}</tr></thead>
  <tbody>
  {{range .Rows}}<tr>{{range .}}<td>{{.}}</td>{{end}}</tr>
  {{end}}
  </tbody>
</table>
</body>
</html>`))

func generateHTML(data *ReportData) ([]byte, error) {
	var buf bytes.Buffer
	if err := htmlTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func filterFromReq(req *dto.GenerateReportRequest) dto.DateFilter {
	f := dto.DateFilter{}
	f.AgentID = req.AgentID
	f.Priority = req.Priority
	f.Category = req.Category
	f.Status = req.Status
	f.Source = req.Source

	now := time.Now()
	switch req.Period {
	case "today":
		f.StartDate = truncateDay(now)
		f.EndDate = now
	case "yesterday":
		y := truncateDay(now).AddDate(0, 0, -1)
		f.StartDate = y
		f.EndDate = y.Add(24*time.Hour - time.Second)
	case "last7":
		f.StartDate = truncateDay(now).AddDate(0, 0, -7)
		f.EndDate = now
	case "last30":
		f.StartDate = truncateDay(now).AddDate(0, 0, -30)
		f.EndDate = now
	case "last90":
		f.StartDate = truncateDay(now).AddDate(0, 0, -90)
		f.EndDate = now
	case "custom":
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			f.StartDate = t
		}
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			f.EndDate = t.Add(24*time.Hour - time.Second)
		}
	default:
		f.StartDate = truncateDay(now).AddDate(0, 0, -30)
		f.EndDate = now
	}
	return f
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
