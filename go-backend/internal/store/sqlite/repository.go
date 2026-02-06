package sqlite

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

type User struct {
	ID            int64
	User          string
	Pwd           string
	RoleID        int
	ExpTime       int64
	Flow          int64
	InFlow        int64
	OutFlow       int64
	FlowResetTime int64
	Num           int
	CreatedTime   int64
	UpdatedTime   sql.NullInt64
	Status        int
}

type ViteConfig struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Time  int64  `json:"time"`
}

type UserTunnelDetail struct {
	ID            int64
	UserID        int64
	TunnelID      int64
	TunnelName    string
	TunnelFlow    int
	Flow          int64
	InFlow        int64
	OutFlow       int64
	Num           int
	FlowResetTime int64
	ExpTime       int64
	SpeedID       sql.NullInt64
	SpeedLimit    sql.NullString
	Speed         sql.NullInt64
}

type UserForwardDetail struct {
	ID         int64
	Name       string
	TunnelID   int64
	TunnelName string
	InIP       string
	InPort     sql.NullInt64
	RemoteAddr string
	InFlow     int64
	OutFlow    int64
	Status     int
	CreatedAt  int64
}

type StatisticsFlow struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Flow      int64  `json:"flow"`
	TotalFlow int64  `json:"totalFlow"`
	Time      string `json:"time"`
}

type Node struct {
	ID      int64
	Secret  string
	Version sql.NullString
	HTTP    int
	TLS     int
	Socks   int
	Status  int
}

func Open(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *Repository) GetUserByUsername(username string) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`
		SELECT id, user, pwd, role_id, exp_time, flow, in_flow, out_flow, flow_reset_time, num, created_time, updated_time, status
		FROM user WHERE user = ? LIMIT 1
	`, username)
	user := &User{}
	if err := row.Scan(
		&user.ID, &user.User, &user.Pwd, &user.RoleID, &user.ExpTime,
		&user.Flow, &user.InFlow, &user.OutFlow, &user.FlowResetTime,
		&user.Num, &user.CreatedTime, &user.UpdatedTime, &user.Status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetConfigByName(name string) (*ViteConfig, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`SELECT id, name, value, time FROM vite_config WHERE name = ? LIMIT 1`, name)
	cfg := &ViteConfig{}
	if err := row.Scan(&cfg.ID, &cfg.Name, &cfg.Value, &cfg.Time); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Repository) ListConfigs() (map[string]string, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	rows, err := r.db.Query(`SELECT name, value FROM vite_config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		result[name] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) UpsertConfig(name, value string, now int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}

	_, err := r.db.Exec(`
		INSERT INTO vite_config(name, value, time)
		VALUES(?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET value=excluded.value, time=excluded.time
	`, name, value, now)
	return err
}

func (r *Repository) GetUserByID(id int64) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`
		SELECT id, user, pwd, role_id, exp_time, flow, in_flow, out_flow, flow_reset_time, num, created_time, updated_time, status
		FROM user WHERE id = ? LIMIT 1
	`, id)
	user := &User{}
	if err := row.Scan(
		&user.ID, &user.User, &user.Pwd, &user.RoleID, &user.ExpTime,
		&user.Flow, &user.InFlow, &user.OutFlow, &user.FlowResetTime,
		&user.Num, &user.CreatedTime, &user.UpdatedTime, &user.Status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *Repository) UsernameExistsExceptID(username string, exceptID int64) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`SELECT COUNT(1) FROM user WHERE user = ? AND id != ?`, username, exceptID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) UpdateUserNameAndPassword(userID int64, username, passwordMD5 string, now int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	_, err := r.db.Exec(`UPDATE user SET user = ?, pwd = ?, updated_time = ? WHERE id = ?`, username, passwordMD5, now, userID)
	return err
}

func (r *Repository) GetUserPackageTunnels(userID int64) ([]UserTunnelDetail, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	rows, err := r.db.Query(`
		SELECT ut.id, ut.user_id, ut.tunnel_id, t.name, t.flow, ut.flow, ut.in_flow, ut.out_flow,
		       ut.num, ut.flow_reset_time, ut.exp_time, ut.speed_id, sl.name, sl.speed
		FROM user_tunnel ut
		LEFT JOIN tunnel t ON t.id = ut.tunnel_id
		LEFT JOIN speed_limit sl ON sl.id = ut.speed_id
		WHERE ut.user_id = ?
		ORDER BY ut.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UserTunnelDetail, 0)
	for rows.Next() {
		var item UserTunnelDetail
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.TunnelID, &item.TunnelName, &item.TunnelFlow,
			&item.Flow, &item.InFlow, &item.OutFlow, &item.Num, &item.FlowResetTime,
			&item.ExpTime, &item.SpeedID, &item.SpeedLimit, &item.Speed,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) GetUserPackageForwards(userID int64) ([]UserForwardDetail, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	rows, err := r.db.Query(`
		SELECT f.id, f.name, f.tunnel_id, t.name, f.remote_addr, f.in_flow, f.out_flow, f.status, f.created_time,
		       GROUP_CONCAT(n.server_ip || ':' || fp.port), MIN(fp.port)
		FROM forward f
		LEFT JOIN tunnel t ON t.id = f.tunnel_id
		LEFT JOIN forward_port fp ON fp.forward_id = f.id
		LEFT JOIN node n ON n.id = fp.node_id
		WHERE f.user_id = ?
		GROUP BY f.id
		ORDER BY f.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UserForwardDetail, 0)
	for rows.Next() {
		var item UserForwardDetail
		if err := rows.Scan(
			&item.ID, &item.Name, &item.TunnelID, &item.TunnelName, &item.RemoteAddr,
			&item.InFlow, &item.OutFlow, &item.Status, &item.CreatedAt, &item.InIP, &item.InPort,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) GetStatisticsFlows(userID int64, limit int) ([]StatisticsFlow, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	rows, err := r.db.Query(`
		SELECT id, user_id, flow, total_flow, time
		FROM statistics_flow
		WHERE user_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StatisticsFlow, 0)
	for rows.Next() {
		var item StatisticsFlow
		if err := rows.Scan(&item.ID, &item.UserID, &item.Flow, &item.TotalFlow, &item.Time); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) NodeExistsBySecret(secret string) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`SELECT COUNT(1) FROM node WHERE secret = ?`, secret)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetNodeBySecret(secret string) (*Node, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("repository not initialized")
	}

	row := r.db.QueryRow(`SELECT id, secret, version, http, tls, socks, status FROM node WHERE secret = ? LIMIT 1`, secret)
	var n Node
	if err := row.Scan(&n.ID, &n.Secret, &n.Version, &n.HTTP, &n.TLS, &n.Socks, &n.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (r *Repository) UpdateNodeOnline(nodeID int64, status int, version string, httpVal, tlsVal, socksVal int) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	_, err := r.db.Exec(`UPDATE node SET status = ?, version = ?, http = ?, tls = ?, socks = ?, updated_time = ? WHERE id = ?`,
		status, version, httpVal, tlsVal, socksVal, unixMilliNow(), nodeID)
	return err
}

func (r *Repository) UpdateNodeStatus(nodeID int64, status int) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}
	_, err := r.db.Exec(`UPDATE node SET status = ?, updated_time = ? WHERE id = ?`, status, unixMilliNow(), nodeID)
	return err
}

func (r *Repository) AddFlow(forwardID, userID int64, userTunnelID int64, inFlow, outFlow int64) error {
	if r == nil || r.db == nil {
		return errors.New("repository not initialized")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`UPDATE forward SET in_flow = in_flow + ?, out_flow = out_flow + ? WHERE id = ?`, inFlow, outFlow, forwardID); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE user SET in_flow = in_flow + ?, out_flow = out_flow + ? WHERE id = ?`, inFlow, outFlow, userID); err != nil {
		return err
	}
	if userTunnelID > 0 {
		if _, err = tx.Exec(`UPDATE user_tunnel SET in_flow = in_flow + ?, out_flow = out_flow + ? WHERE id = ?`, inFlow, outFlow, userTunnelID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func unixMilliNow() int64 {
	return time.Now().UnixMilli()
}
