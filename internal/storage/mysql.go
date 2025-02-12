package storage

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &MySQLStorage{db: db}, nil
}

func (s *MySQLStorage) Close() error {
	return s.db.Close()
}

func (s *MySQLStorage) SaveMockUrl(ctx context.Context, url, respCode, respHeader, respBody, owner, desc string) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO stub_interface (url, def_resp_code, def_resp_header, def_resp_body, owner, desc, status)
         VALUES (?, ?, ?, ?, ?, ?, 'active')`,
		url, respCode, respHeader, respBody, owner, desc)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *MySQLStorage) SaveRule(ctx context.Context, interfaceID int64, matchType int32, matchRule, respCode, respHeader, respBody string, delayTime int32, desc string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO stub_rule (interface_id, match_type, match_rule, resp_code, resp_header, resp_body, delay_time, desc, status)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')`,
		interfaceID, matchType, matchRule, respCode, respHeader, respBody, delayTime, desc)
	return err
}

func (s *MySQLStorage) GetMockResponse(ctx context.Context, url string) (*MockResponse, error) {
	var resp MockResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT id, def_resp_code, def_resp_header, def_resp_body 
         FROM stub_interface 
         WHERE url = ? AND status = 'active'`,
		url).Scan(&resp.InterfaceID, &resp.ResponseCode, &resp.ResponseHeader, &resp.ResponseBody)
	return &resp, err
}

type MockResponse struct {
	InterfaceID    int64
	ResponseCode   string
	ResponseHeader string
	ResponseBody   string
}

type Rule struct {
	MatchType      int32
	MatchRule      string
	ResponseCode   string
	ResponseHeader string
	ResponseBody   string
	DelayTime      int32
}

func (s *MySQLStorage) GetRules(ctx context.Context, interfaceID int64) ([]Rule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT match_type, match_rule, resp_code, resp_header, resp_body, delay_time 
         FROM stub_rule 
         WHERE interface_id = ? AND status = 'active'`,
		interfaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.MatchType, &rule.MatchRule, &rule.ResponseCode, &rule.ResponseHeader, &rule.ResponseBody, &rule.DelayTime); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
