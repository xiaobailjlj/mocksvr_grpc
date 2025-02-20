package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/model"
	"github.com/xiaobailjlj/mocksvr_grpc/internal/pkg/logger"
	"go.uber.org/zap"
	"time"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Error("Failed to open database connection",
			zap.String("dsn", dsn),
			zap.Error(err))
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database",
			zap.String("dsn", dsn),
			zap.Error(err))
		return nil, err
	}

	return &MySQLStorage{db: db}, nil
}

func (s *MySQLStorage) Close() error {
	if err := s.db.Close(); err != nil {
		logger.Error("Failed to close database connection",
			zap.Error(err))
		return err
	}
	return nil
}

func (s *MySQLStorage) SaveMockUrl(ctx context.Context, url, respCode string, respHeader map[string]string, respBody, owner, description, meta string) (int64, error) {
	start := time.Now()

	// Convert header map to JSON string
	headerJSON, err := json.Marshal(respHeader)
	if err != nil {
		logger.Error("Failed to marshal response header",
			zap.Any("header", respHeader),
			zap.Error(err))
		return 0, fmt.Errorf("failed to marshal response header: %v", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction",
			zap.Error(err))
		return 0, err
	}
	defer tx.Rollback()

	// First, try to get existing ID
	var existingID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM stub_interface WHERE url = ?", url).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("Failed to query existing interface",
			zap.String("url", url),
			zap.Error(err))
		return 0, fmt.Errorf("failed to query existing interface: %v", err)
	}

	query := `INSERT INTO stub_interface (
        url, def_resp_code, def_resp_header, def_resp_body, 
        owner, description, meta, status
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
        def_resp_code = VALUES(def_resp_code),
        def_resp_header = VALUES(def_resp_header),
        def_resp_body = VALUES(def_resp_body),
        owner = VALUES(owner),
        description = VALUES(description),
        meta = VALUES(meta),
        status = VALUES(status)`

	// Insert or update stub_interface
	result, err := tx.ExecContext(ctx, query,
		url, respCode, string(headerJSON), respBody,
		owner, description, meta, model.StatusActive)
	if err != nil {
		logger.Error("Failed to insert stub interface",
			zap.String("query", query),
			zap.String("url", url),
			zap.String("respCode", respCode),
			zap.String("owner", owner),
			zap.Error(err))
		return 0, fmt.Errorf("failed to insert stub interface: %v", err)
	}

	// Get the ID - use existing ID if it was an update
	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		id = existingID // Use the existing ID if this was an update
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction",
			zap.Error(err))
		return 0, err
	}

	logger.Info("Successfully saved mock URL",
		zap.Int64("id", id),
		zap.String("url", url),
		zap.Bool("isUpdate", existingID > 0),
		zap.Duration("duration", time.Since(start)))

	return id, nil
}

func (s *MySQLStorage) SaveRule(ctx context.Context, interfaceID int64, rule *model.Rule) error {
	start := time.Now()

	headerJSON, err := json.Marshal(rule.ResponseHeader)
	if err != nil {
		logger.Error("Failed to marshal rule response header",
			zap.Any("header", rule.ResponseHeader),
			zap.Error(err))
		return fmt.Errorf("failed to marshal rule response header: %v", err)
	}

	logger.Debug("SaveRule",
		zap.Int64("interfaceID", interfaceID),
		zap.Int32("matchType", rule.MatchType),
		zap.Error(err))

	// SaveRule query
	query := `INSERT INTO stub_rule (
    interface_id, match_type, match_rule, 
    resp_code, resp_header, resp_body,
    delay_time, description, meta, status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    match_rule = VALUES(match_rule),
    resp_code = VALUES(resp_code),
    resp_header = VALUES(resp_header),
    resp_body = VALUES(resp_body),
    delay_time = VALUES(delay_time),
    description = VALUES(description),
    meta = VALUES(meta),
    status = VALUES(status)`

	_, err = s.db.ExecContext(ctx, query,
		interfaceID, rule.MatchType, rule.MatchRule,
		rule.ResponseCode, string(headerJSON), rule.ResponseBody,
		rule.DelayTime, rule.Description, rule.Meta, model.StatusActive)

	if err != nil {
		logger.Error("Failed to insert rule",
			zap.String("query", query),
			zap.Int64("interfaceID", interfaceID),
			zap.Int32("matchType", rule.MatchType),
			zap.Error(err))
		return fmt.Errorf("failed to insert rule: %v", err)
	}

	logger.Info("Successfully saved rule",
		zap.Int64("interfaceID", interfaceID),
		zap.Int32("matchType", rule.MatchType),
		zap.Duration("duration", time.Since(start)))

	return nil
}

func (s *MySQLStorage) GetMockResponse(ctx context.Context, url string) (*model.MockResponse, error) {
	start := time.Now()

	var resp model.MockResponse
	var headerJSON string

	query := `SELECT id, def_resp_code, def_resp_header, def_resp_body 
		FROM stub_interface 
		WHERE url = ? AND status = ?`

	err := s.db.QueryRowContext(ctx, query,
		url, model.StatusActive).Scan(&resp.InterfaceID, &resp.ResponseCode, &headerJSON, &resp.ResponseBody)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("No mock response found",
				zap.String("url", url))
		} else {
			logger.Error("Failed to get mock response",
				zap.String("query", query),
				zap.String("url", url),
				zap.Error(err))
		}
		return nil, err
	}

	// Parse header JSON
	if headerJSON != "" {
		if err := json.Unmarshal([]byte(headerJSON), &resp.ResponseHeader); err != nil {
			logger.Error("Failed to unmarshal header JSON",
				zap.String("headerJSON", headerJSON),
				zap.Error(err))
			return nil, fmt.Errorf("failed to unmarshal header JSON: %v", err)
		}
	}

	logger.Debug("Successfully retrieved mock response",
		zap.String("url", url),
		zap.Int64("interfaceID", resp.InterfaceID),
		zap.Duration("duration", time.Since(start)))

	return &resp, nil
}

func (s *MySQLStorage) GetRules(ctx context.Context, interfaceID int64) ([]model.Rule, error) {
	start := time.Now()

	query := `SELECT match_type, match_rule, resp_code, resp_header, resp_body, delay_time, description, meta 
		FROM stub_rule 
		WHERE interface_id = ? AND status = ?`

	rows, err := s.db.QueryContext(ctx, query,
		interfaceID, model.StatusActive)
	if err != nil {
		logger.Error("Failed to query rules",
			zap.String("query", query),
			zap.Int64("interfaceID", interfaceID),
			zap.Error(err))
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
			logger.Error("Failed to scan rule row",
				zap.Error(err))
			return nil, err
		}

		// Parse header JSON
		if headerJSON != "" {
			if err := json.Unmarshal([]byte(headerJSON), &rule.ResponseHeader); err != nil {
				logger.Error("Failed to unmarshal rule header JSON",
					zap.String("headerJSON", headerJSON),
					zap.Error(err))
				return nil, fmt.Errorf("failed to unmarshal rule header JSON: %v", err)
			}
		}

		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error after iterating rules rows",
			zap.Error(err))
		return nil, err
	}

	logger.Debug("Successfully retrieved rules",
		zap.Int64("interfaceID", interfaceID),
		zap.Int("count", len(rules)),
		zap.Duration("duration", time.Since(start)))

	return rules, nil
}

func (s *MySQLStorage) GetAllMockUrls(ctx context.Context, owner string, page, pageSize int) ([]*model.Interface, int, error) {
	start := time.Now()

	// Calculate offset
	offset := (page - 1) * pageSize

	// Base query
	baseQuery := `SELECT 
        id, url, def_resp_code, def_resp_header, def_resp_body, 
        owner, description, meta
    FROM stub_interface 
    WHERE status = ?`
	countQuery := `SELECT COUNT(*) FROM stub_interface WHERE status = ?`

	args := []interface{}{model.StatusActive}

	// Add owner filter if provided
	if owner != "" {
		baseQuery += " AND owner = ?"
		countQuery += " AND owner = ?"
		args = append(args, owner)
	}

	// Add pagination
	baseQuery += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		logger.Error("Failed to get total count",
			zap.String("query", countQuery),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total count: %v", err)
	}

	// Execute main query
	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		logger.Error("Failed to query mock URLs",
			zap.String("query", baseQuery),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query mock URLs: %v", err)
	}
	defer rows.Close()

	var interfaces []*model.Interface
	for rows.Next() {
		var iface model.Interface
		var headerJSON string

		err := rows.Scan(
			&iface.ID,
			&iface.URL,
			&iface.ResponseCode,
			&headerJSON,
			&iface.ResponseBody,
			&iface.Owner,
			&iface.Description,
			&iface.Meta,
		)
		if err != nil {
			logger.Error("Failed to scan mock URL row",
				zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan mock URL row: %v", err)
		}

		// Parse header JSON
		if headerJSON != "" {
			if err := json.Unmarshal([]byte(headerJSON), &iface.ResponseHeader); err != nil {
				logger.Error("Failed to unmarshal response header",
					zap.String("header", headerJSON),
					zap.Error(err))
				return nil, 0, fmt.Errorf("failed to unmarshal response header: %v", err)
			}
		}

		interfaces = append(interfaces, &iface)
	}

	logger.Info("Successfully retrieved mock URLs",
		zap.Int("count", len(interfaces)),
		zap.Int("total", total),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
		zap.Duration("duration", time.Since(start)))

	return interfaces, total, nil
}

func (s *MySQLStorage) GetRulesByInterfaceID(ctx context.Context, interfaceID int64) ([]*model.Rule, error) {
	start := time.Now()

	query := `SELECT 
        match_type, match_rule, 
        resp_code, resp_header, resp_body,
        delay_time, description, meta
    FROM stub_rule 
    WHERE interface_id = ? AND status = ?
    ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, query, interfaceID, model.StatusActive)
	if err != nil {
		logger.Error("Failed to query rules",
			zap.String("query", query),
			zap.Int64("interfaceID", interfaceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query rules: %v", err)
	}
	defer rows.Close()

	var rules []*model.Rule
	for rows.Next() {
		var rule model.Rule
		var headerJSON string

		err := rows.Scan(
			//&rule.ID,
			//&rule.interfaceID,
			&rule.MatchType,
			&rule.MatchRule,
			&rule.ResponseCode,
			&headerJSON,
			&rule.ResponseBody,
			&rule.DelayTime,
			&rule.Description,
			&rule.Meta,
		)
		if err != nil {
			logger.Error("Failed to scan rule row",
				zap.Error(err))
			return nil, fmt.Errorf("failed to scan rule row: %v", err)
		}

		// Parse header JSON
		if headerJSON != "" {
			if err := json.Unmarshal([]byte(headerJSON), &rule.ResponseHeader); err != nil {
				logger.Error("Failed to unmarshal rule response header",
					zap.String("header", headerJSON),
					zap.Error(err))
				return nil, fmt.Errorf("failed to unmarshal rule response header: %v", err)
			}
		}

		rules = append(rules, &rule)
	}

	logger.Info("Successfully retrieved rules",
		zap.Int64("interfaceID", interfaceID),
		zap.Int("count", len(rules)),
		zap.Duration("duration", time.Since(start)))

	return rules, nil
}
