package db

import (
	"fmt"
	"strings"
)

type SelectQueryBuilder struct {
	Columns string
	From    string
	Where   string
	GroupBy string
	OrderBy string
}

type InsertQueryBuilder struct {
	Table       string
	Columns     string
	ReturnValue string
}

type UpdateQueryBuilder struct {
	Table string
	Set   string
	Where string
}

type DeleteQueryBuilder struct {
	Table string
	Where string
}

func (s *SelectQueryBuilder) Build() string {
	query := "SELECT "

	// Add columns
	if s.Columns != "" {
		query += s.Columns
	} else {
		query += "*"
	}

	// Add FROM clause
	if s.From != "" {
		query += " FROM " + s.From
	}

	// Add WHERE clause
	if s.Where != "" {
		query += " WHERE " + s.Where
	}

	// Add GROUP BY clause
	if s.GroupBy != "" {
		query += " GROUP BY " + s.GroupBy
	}

	// Add ORDER BY clause
	if s.OrderBy != "" {
		query += " ORDER BY " + s.OrderBy
	}

	return query
}

func (s *InsertQueryBuilder) Build() string {
	// Split the columns by commas to count the number of columns
	columnList := strings.Split(s.Columns, ",")
	numValues := len(columnList)

	// Generate placeholders ($1, $2, ..., $n) based on the number of columns
	placeholders := make([]string, numValues)
	for i := 0; i < numValues; i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Join the placeholders with commas
	values := strings.Join(placeholders, ", ")

	// Start building the query
	query := "INSERT INTO " + s.Table + " (" + s.Columns + ") VALUES (" + values + ")"

	// Add RETURNING clause if returnValue is specified
	if s.ReturnValue != "" {
		query += " RETURNING " + s.ReturnValue
	}

	return query
}

func (u *UpdateQueryBuilder) Build() string {
	// Start building the query
	query := "UPDATE " + u.Table + " SET " + u.Set

	// Add WHERE clause if specified
	if u.Where != "" {
		query += " WHERE " + u.Where
	}

	return query
}

func (d *DeleteQueryBuilder) Build() string {
	// Start building the DELETE query
	query := "DELETE FROM " + d.Table

	// Add WHERE clause if specified
	if d.Where != "" {
		query += " WHERE " + d.Where
	}

	return query
}
