package pdcpserver

import (
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

func New(address string, port uint32) PdcpServer {
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
