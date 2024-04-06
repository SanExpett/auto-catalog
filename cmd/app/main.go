package main

import (
	"fmt"
	"github.com/SanExpett/auto-catalog/internal/server"
	"github.com/SanExpett/auto-catalog/pkg/config"
)

//	@title      AUTO-CATALOG project API
//	@version    1.0
//	@description  This is a server of AUTO-CATALOG server.
//
// @Schemes http
// @BasePath  /api/v1
func main() {
	configServer := config.New()

	srv := new(server.Server)
	if err := srv.Run(configServer); err != nil {
		fmt.Printf("Error in server: %s", err.Error())
	}
}
