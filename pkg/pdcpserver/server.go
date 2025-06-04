package pdcpserver

import (
	"log/slog"
	"pdcplet/pkg/pdcpserver/database"
	"pdcplet/pkg/pdcpserver/router"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PdcpServer interface {
	Start() error
}

type pdcpServer struct {
	address string
	port    uint32
	engine  *gin.Engine
}

func New(address string, port uint32, sqlite3Path string) PdcpServer {

	if err := database.InitSQLite(sqlite3Path); err != nil {
		slog.Error("Failed to init sqlite3 database", "error", err, "sqlite3Path", sqlite3Path)
		panic(err)
	}

	r := gin.Default()

	router.RegisterVMRoutes(r)

	return &pdcpServer{
		address: address,
		port:    port,
		engine:  r,
	}
}

func (s *pdcpServer) Start() error {
	return s.engine.Run(s.address + ":" + strconv.FormatUint(uint64(s.port), 10))
}
