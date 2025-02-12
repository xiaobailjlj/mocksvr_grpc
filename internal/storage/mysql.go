package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/model"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLStorage{db: db}, nil
}

func (s *MySQLStorage) Close() error {
	return s.db.Close()
}

func (s *MySQLStorage) SaveMockUrl(ctx context.Context, url, respCode string, respHeader map[string]string, respBody, owner, desc, meta string) (int64, error) {
	// Convert header map to JSON string
	headerJSON, err := json.Marshal(respHeader)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal response header: %v", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert into stub_interface
	result, err := tx.ExecContext(ctx,
		`INSERT INTO stub_interface (
            url, def_resp_code, def_resp_header, def_resp_body, 
            owner, desc, meta, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		url, respCode, string(headerJSON), respBody,
		owner, desc, meta, model.StatusActive)
	if err != nil {
		return 0, fmt.Errorf("failed to insert stub interface: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *MySQLStorage) SaveRule(ctx context.Context, interfaceID int64, rule *model.Rule) error {
	headerJSON, err := json.Marshal(rule.ResponseHeader)
	if err != nil {
		return fmt.Errorf("failed to marshal rule response header: %v", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO stub_rule (
            interface_id, match_type, match_rule, 
            resp_code, resp_header, resp_body,
            delay_time, desc, meta, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		interfaceID, rule.MatchType, rule.MatchRule,
		rule.ResponseCode, string(headerJSON), rule.ResponseBody,
		rule.DelayTime, rule.Description, rule.Meta, model.StatusActive)

	if err != nil {
		return fmt.Errorf("failed to insert rule: %v", err)
	}

	return nil
}

func (s *MySQLStorage) GetMockResponse(ctx context.Context, url string) (*model.MockResponse, error) {
	var resp model.MockResponse
	var headerJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, def_resp_code, def_resp_header, def_resp_body 
         FROM stub_interface 
         WHERE url = ? AND status = ?`,
		url, model.StatusActive).Scan(&resp.InterfaceID, &resp.ResponseCode, &headerJSON, &resp.ResponseBody)
	if err != nil {
		return nil, err
	}

	// Parse header JSON
	if headerJSON != "" {
		if err := json.Unmarshal([]byte(headerJSON), &resp.ResponseHeader); err != nil {
			return nil, fmt.Errorf("failed to unmarshal header JSON: %v", err)
		}
	}

	return &resp, nil
}

func (s *MySQLStorage) GetRules(ctx context.Context, interfaceID int64) ([]model.Rule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT match_type, match_rule, resp_code, resp_header, resp_body, delay_time, desc, meta 
         FROM stub_rule 
         WHERE interface_id = ? AND status = ?`,
		interfaceID, model.StatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.Rule
	for rows.Next() {
		var rule model.Rule
		var headerJSON string
		if err := rows.Scan(
			&rule.MatchType, &rule.MatchRule, &rule.ResponseCode,
			&headerJSON, &rule.ResponseBody, &rule.DelayTime,
			&rule.Description, &rule.Meta,
		); err != nil {
			return nil, err
		}

		// Parse header JSON
		if headerJSON != "" {
			if err := json.Unmarshal([]byte(headerJSON), &rule.ResponseHeader); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rule header JSON: %v", err)
			}
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

type MockResponse struct {
	InterfaceID    int64             `json:"interface_id"`
	ResponseCode   string            `json:"response_code"`
	ResponseHeader map[string]string `json:"response_header"`
	ResponseBody   string            `json:"response_body"`
}
