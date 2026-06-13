// Package pagination provides cursor-based pagination helpers for PostgreSQL
// queries via pgx. Cursor pagination is preferred over offset pagination for
// large datasets because it remains stable as rows are inserted/deleted.
package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const defaultPageSize = 20
const maxPageSize = 100

// Page holds the decoded pagination parameters for a query.
type Page struct {
	// After is the cursor value to start after (exclusive). Empty means start
	// from the beginning.
	After string
	// Limit is the maximum number of rows to return.
	Limit int
}

// PageFromQuery decodes cursor and limit from URL query parameters.
//
//	after := r.URL.Query().Get("after")
//	limitStr := r.URL.Query().Get("limit")
//	page, err := pagination.PageFromQuery(after, limitStr)
func PageFromQuery(after, limitStr string) (Page, error) {
	p := Page{Limit: defaultPageSize}

	if after != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(after)
		if err != nil {
			return p, fmt.Errorf("pagination: invalid cursor: %w", err)
		}
		p.After = string(decoded)
	}

	if limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n < 1 {
			return p, fmt.Errorf("pagination: limit must be a positive integer")
		}
		if n > maxPageSize {
			n = maxPageSize
		}
		p.Limit = n
	}

	return p, nil
}

// EncodeCursor encodes a row's sort key as a base64url cursor string.
// Pass the last row's ID or sort field as the value.
func EncodeCursor(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

// Result wraps a page of items with the next cursor.
type Result[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// NewResult builds a paginated result from a slice.
// Pass one extra item beyond the page limit to detect whether more pages exist.
// The extra item is trimmed from Items and used to generate NextCursor.
//
//	rows, err := repo.List(ctx, page.Limit+1, page.After)
//	return pagination.NewResult(rows, page.Limit, func(r Row) string { return r.ID })
func NewResult[T any](items []T, limit int, cursorKey func(T) string) Result[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	r := Result[T]{
		Items:   items,
		HasMore: hasMore,
	}
	if hasMore && len(items) > 0 {
		r.NextCursor = EncodeCursor(cursorKey(items[len(items)-1]))
	}
	return r
}

// IDCursorSQL returns a SQL WHERE clause fragment for cursor-based pagination
// on an integer ID column, e.g. "AND id > $N".
// Pass the parameter index for the cursor value.
func IDCursorSQL(after string, paramIdx int) (clause string, args []interface{}) {
	if after == "" {
		return "", nil
	}
	// Cursor for IDs is stored as the string representation of the ID.
	parts := strings.SplitN(after, ":", 2)
	id := parts[0]
	return fmt.Sprintf("AND id > $%d", paramIdx), []interface{}{id}
}
