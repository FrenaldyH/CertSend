package main

import (
	"CertSend/internal/service"
	"CertSend/pkg/logger"
	"context"
	"fmt"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := logger.InitLog("Logs"); err != nil {
		println("failed to init logger:", err.Error())
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SendCertificates(certPath string, csvPath string, host string, port int, username string, password string, delaySeconds int) (error) {
	smtp := service.SMTPConfig{
		Host: host,
		Port: port,
		Username: username,
		Password: password,
	}
	return service.SendCertificates(
		certPath,
		csvPath,
		smtp,
		time.Duration(delaySeconds) * time.Second,
	)
}
