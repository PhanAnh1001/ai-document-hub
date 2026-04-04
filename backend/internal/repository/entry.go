package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

type EntryRepository struct {
	pool *pgxpool.Pool
}

func NewEntryRepository(pool *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{pool: pool}
}

// CountByStatus returns a map of status → count for all entries in the org.
func (r *EntryRepository) CountByStatus(ctx context.Context, orgID string) (map[string]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM accounting_entries
		 WHERE organization_id=$1 GROUP BY status`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("count entries by status: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan count row: %w", err)
		}
		counts[status] = count
	}
	return counts, nil
}

// SumApprovedAmount returns the total amount of approved entries for an organization.
func (r *EntryRepository) SumApprovedAmount(ctx context.Context, orgID string) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM accounting_entries
		 WHERE organization_id=$1 AND status='approved'`,
		orgID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum approved amount: %w", err)
	}
	return total, nil
}

// CountByDay returns the number of entries created per calendar day for the past N days.
func (r *EntryRepository) CountByDay(ctx context.Context, orgID string, days int) ([]model.DayCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT TO_CHAR(entry_date::date, 'YYYY-MM-DD') AS day, COUNT(*)::int
		 FROM accounting_entries
		 WHERE organization_id=$1
		   AND entry_date >= CURRENT_DATE - ($2::int - 1) * INTERVAL '1 day'
		 GROUP BY day
		 ORDER BY day ASC`,
		orgID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("count entries by day: %w", err)
	}
	defer rows.Close()

	var counts []model.DayCount
	for rows.Next() {
		var dc model.DayCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, fmt.Errorf("scan day count: %w", err)
		}
		counts = append(counts, dc)
	}
	return counts, nil
}

// CreateBatch inserts multiple accounting entries in a single transaction.
func (r *EntryRepository) CreateBatch(ctx context.Context, entries []*model.AccountingEntry) error {
	if len(entries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, e := range entries {
		batch.Queue(
			`INSERT INTO accounting_entries
			 (id, organization_id, document_id, entry_date, description,
			  debit_account, credit_account, amount, status, ai_confidence)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			e.ID, e.OrganizationID, e.DocumentID, e.EntryDate, e.Description,
			e.DebitAccount, e.CreditAccount, e.Amount, e.Status, e.AIConfidence,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range entries {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert entry: %w", err)
		}
	}
	return nil
}

// allowedSortColumns is a whitelist of SQL columns for ORDER BY to prevent injection.
var allowedSortColumns = map[string]string{
	"entry_date":  "ae.entry_date",
	"amount":      "ae.amount",
	"created_at":  "ae.created_at",
	"description": "ae.description",
}

// ListByOrg returns entries for an organization, optionally filtered by status, search query, date range, and sorted.
func (r *EntryRepository) ListByOrg(ctx context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error) {
	var (
		rows  pgx.Rows
		err   error
		total int
	)

	// Build WHERE clauses dynamically (using ae. alias for the JOIN query)
	where := "WHERE ae.organization_id=$1"
	args := []interface{}{orgID}
	idx := 2

	if status != "" {
		where += fmt.Sprintf(" AND ae.status=$%d", idx)
		args = append(args, status)
		idx++
	}
	if q != "" {
		where += fmt.Sprintf(" AND ae.description ILIKE '%%' || $%d || '%%'", idx)
		args = append(args, q)
		idx++
	}
	if dateFrom != "" {
		where += fmt.Sprintf(" AND ae.entry_date >= $%d", idx)
		args = append(args, dateFrom)
		idx++
	}
	if dateTo != "" {
		where += fmt.Sprintf(" AND ae.entry_date <= $%d", idx)
		args = append(args, dateTo)
		idx++
	}

	err = r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM accounting_entries ae
		 LEFT JOIN documents d ON d.id = ae.document_id %s`, where),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count entries: %w", err)
	}

	countArgs := args
	args = append(args, limit, offset)
	_ = countArgs

	// Build ORDER BY — whitelist column names to prevent SQL injection
	orderCol := "ae.created_at"
	if col, ok := allowedSortColumns[sortBy]; ok {
		orderCol = col
	}
	orderDir := "DESC"
	if strings.EqualFold(sortDir, "asc") {
		orderDir = "ASC"
	}

	rows, err = r.pool.Query(ctx,
		fmt.Sprintf(`SELECT ae.id, ae.organization_id, ae.document_id, d.file_name,
		        ae.entry_date, ae.description,
		        ae.debit_account, ae.credit_account, ae.amount, ae.status,
		        ae.reject_reason, ae.ai_confidence,
		        ae.created_at, ae.updated_at
		 FROM accounting_entries ae
		 LEFT JOIN documents d ON d.id = ae.document_id
		 %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, orderCol, orderDir, idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	var entries []*model.AccountingEntry
	for rows.Next() {
		var e model.AccountingEntry
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.DocumentID, &e.DocumentName, &e.EntryDate, &e.Description,
			&e.DebitAccount, &e.CreditAccount, &e.Amount, &e.Status,
			&e.RejectReason, &e.AIConfidence, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan entry: %w", err)
		}
		entries = append(entries, &e)
	}
	return entries, total, nil
}

// BulkUpdateStatus updates status for multiple entries scoped to the organization in one query.
func (r *EntryRepository) BulkUpdateStatus(ctx context.Context, ids []string, orgID string, status model.EntryStatus) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	// Build $2,$3,... placeholders for id list
	args := []interface{}{status, orgID}
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args = append(args, id)
		placeholders[i] = fmt.Sprintf("$%d", i+3)
	}
	query := fmt.Sprintf(
		"UPDATE accounting_entries SET status=$1, updated_at=NOW() WHERE organization_id=$2 AND id IN (%s)",
		strings.Join(placeholders, ","),
	)
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("bulk update entry status: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// UpdateStatus changes the status of an entry scoped to the organization.
func (r *EntryRepository) UpdateStatus(ctx context.Context, id, orgID string, status model.EntryStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE accounting_entries SET status=$1 WHERE id=$2 AND organization_id=$3`,
		status, id, orgID,
	)
	if err != nil {
		return fmt.Errorf("update entry status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found")
	}
	return nil
}

// UpdateStatusWithReason changes the status of an entry and stores a rejection reason.
func (r *EntryRepository) UpdateStatusWithReason(ctx context.Context, id, orgID string, status model.EntryStatus, reason string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE accounting_entries SET status=$1, reject_reason=$2, updated_at=NOW()
		 WHERE id=$3 AND organization_id=$4`,
		status, reason, id, orgID,
	)
	if err != nil {
		return fmt.Errorf("update entry status with reason: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found")
	}
	return nil
}

// UpdateEntry updates editable fields of an entry scoped to the organization.
func (r *EntryRepository) UpdateEntry(ctx context.Context, id, orgID string, fields model.UpdateEntryFields) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE accounting_entries
		 SET description=$1, debit_account=$2, credit_account=$3, amount=$4, updated_at=NOW()
		 WHERE id=$5 AND organization_id=$6`,
		fields.Description, fields.DebitAccount, fields.CreditAccount, fields.Amount, id, orgID,
	)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found")
	}
	return nil
}

// UpdateStatusByID updates the status of an entry by ID only (no org scope).
// Used by MISA callback which identifies entries by org_refid (our entry ID).
func (r *EntryRepository) UpdateStatusByID(ctx context.Context, id string, status model.EntryStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE accounting_entries SET status=$1, updated_at=NOW() WHERE id=$2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update entry status by id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found")
	}
	return nil
}

// GetByID fetches a single entry scoped to the organization.
func (r *EntryRepository) GetByID(ctx context.Context, id, orgID string) (*model.AccountingEntry, error) {
	var e model.AccountingEntry
	err := r.pool.QueryRow(ctx,
		`SELECT ae.id, ae.organization_id, ae.document_id, d.file_name,
		        ae.entry_date, ae.description,
		        ae.debit_account, ae.credit_account, ae.amount, ae.status,
		        ae.reject_reason, ae.ai_confidence,
		        ae.created_at, ae.updated_at
		 FROM accounting_entries ae
		 LEFT JOIN documents d ON d.id = ae.document_id
		 WHERE ae.id=$1 AND ae.organization_id=$2`,
		id, orgID,
	).Scan(
		&e.ID, &e.OrganizationID, &e.DocumentID, &e.DocumentName, &e.EntryDate, &e.Description,
		&e.DebitAccount, &e.CreditAccount, &e.Amount, &e.Status,
		&e.RejectReason, &e.AIConfidence, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get entry: %w", err)
	}
	return &e, nil
}
