package repository

import (
	"database/sql"
	"github.com/pawski/proxkeep/domain/proxy"
	"time"
)

type ProxyServerRepository struct {
	db     *sql.DB
	logger proxy.Logger
}

func NewProxyServerRepository(database *sql.DB, logger proxy.Logger) *ProxyServerRepository {
	return &ProxyServerRepository{database, logger}
}

func (r *ProxyServerRepository) FindAll() []proxy.Server {

	servers := make([]proxy.Server, 0)

	query := "SELECT uid, ip, port, is_available, throughoutput_rate, COALESCE(failure_reason, '') FROM proxy_server ORDER BY updated_at ASC"
	rows, err := r.db.Query(query)

	if err != nil {
		r.logger.Errorf("Failed to retrieve ProxyServers. %v", err)
		return servers
	}

	defer rows.Close()

	for rows.Next() {
		var s proxy.Server
		err := rows.Scan(&s.Uid, &s.Ip, &s.Port, &s.IsAvailable, &s.ThroughputRate, &s.FailureReason)

		if err != nil {
			r.logger.Fatal(err)
		}

		servers = append(servers, s)
	}

	if err := rows.Err(); err != nil {
		r.logger.Fatal(err)
	}

	return servers
}

func (r *ProxyServerRepository) FindByUid(uid proxy.Uid) proxy.Server {
	return proxy.Server{}
}

func (r *ProxyServerRepository) Persist(s proxy.Server) error {

	stmt, err := r.db.Prepare("UPDATE proxy_server SET is_available=?, throughoutput_rate=?, failure_reason=?, updated_at=? WHERE uid=?")

	if err != nil {
		r.logger.Errorf("Failed to prepare persist statement for ProxyServer. %v", err)
		return err
	}

	updatedAt := time.Now().UTC()
	res, err := stmt.Exec(s.IsAvailable, s.ThroughputRate, s.FailureReason, updatedAt.Format(time.RFC3339), s.Uid)

	if err != nil {
		r.logger.Errorf("Failed to execute persist statement for ProxyServer. %v", err)
		return err
	}

	_, err = res.LastInsertId()

	if err != nil {
		r.logger.Errorf("Failed to get last insert id for persist of ProxyServer. %v", err)
		return err
	}

	return nil
}
