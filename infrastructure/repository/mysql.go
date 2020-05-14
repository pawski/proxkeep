package repository

import (
	"database/sql"
	"github.com/pawski/proxkeep/logger"
)

type ProxyServerRepository struct {
	db     *sql.DB
	logger logger.Logger
}

type ProxyServer struct {
	Id             uint
	Uid            string
	Ip             string
	Port           string
	IsAvailable    bool
	ThroughputRate float32
}

func NewProxyServerRepository(database *sql.DB, logger logger.Logger) *ProxyServerRepository {
	return &ProxyServerRepository{database, logger}
}

func (r *ProxyServerRepository) FindAll() []ProxyServer {

	servers := make([]ProxyServer, 0)

	rows, err := r.db.Query("SELECT id, uid, ip, port, is_available, throughoutput_rate FROM proxy_server ORDER BY id DESC")

	if err != nil {
		r.logger.Errorf("Failed to retrieve ProxyServers. %v", err)
		return servers
	}

	defer rows.Close()

	for rows.Next() {
		var s ProxyServer
		err := rows.Scan(&s.Id, &s.Uid, &s.Ip, &s.Port, &s.IsAvailable, &s.ThroughputRate)

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
