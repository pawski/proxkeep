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

func (r *ProxyServerRepository) FindAll() []ServerEntity {

	servers := make([]ServerEntity, 0)

	sql := "SELECT id, uid, ip, port, is_available, throughoutput_rate, COALESCE(failure_reason, ''), created_at, updated_at FROM proxy_server ORDER BY updated_at ASC"
	rows, err := r.db.Query(sql)

	if err != nil {
		r.logger.Errorf("Failed to retrieve ProxyServers. %v", err)
		return servers
	}

	defer rows.Close()

	for rows.Next() {
		var s ServerEntity
		err := rows.Scan(&s.Id, &s.Uid, &s.Ip, &s.Port, &s.IsAvailable, &s.ThroughputRate, &s.FailureReason, &s.CreatedAt, &s.UpdatedAt)

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

func (r *ProxyServerRepository) FindByUid(uid proxy.Uid) ServerEntity {
	return ServerEntity{}
}

func (r *ProxyServerRepository) Persist(proxyServer *ServerEntity) error {

	stmt, err := r.db.Prepare("UPDATE proxy_server SET is_available=?, throughoutput_rate=?, failure_reason=?, updated_at=? WHERE uid=?")

	if err != nil {
		r.logger.Errorf("Failed to prepare persist statement for ProxyServer. %v", err)
		return err
	}

	updatedAt := time.Now().UTC()
	res, err := stmt.Exec(proxyServer.IsAvailable, proxyServer.ThroughputRate, proxyServer.FailureReason, updatedAt.Format(time.RFC3339), proxyServer.Uid)

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
