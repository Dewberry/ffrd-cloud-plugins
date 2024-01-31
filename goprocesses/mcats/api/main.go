package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/controller"
	"app/mcats"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

func main() {
	var ctrl *controller.Controller
	var dbErr error
	ctrl, dbErr = controller.NewController()
	log.SetFormatter(&log.JSONFormatter{})
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithError(err).Error("Invalid log level")
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetReportCaller(true)
	log.Infof("level level set to: %s", level)

	// Add retry loop to ensure database connection
	// useful in development if running postgres on docker network
	// and api starts befor database is ready
	if dbErr != nil {
		retryCount := 3
		for i := 0; i < retryCount; i++ {
			ctrl, dbErr = controller.NewController()
			if dbErr == nil {
				break
			}

			log.Infof("Attempt %d: database connection failed: %s\n", i+1, dbErr.Error())

			if i < retryCount-1 {
				time.Sleep(10 * time.Second)
			} else {
				log.Fatalf("unable to connect to database after %v attempts: ", retryCount)
			}
		}
	}

	log.Info("Successfully connected to database!")

	// log.Info("S3 Inventory complete")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
	}))

	e.GET("/ping", ctrl.Ping)
	// // filestore / mcat
	e.GET("/mcat/break_line", mcats.GetGeoJSONHandler(ctrl))
	e.GET("/mcat/mesh_line", mcats.GetGeoJSONHandler(ctrl))
	e.GET("/mcat/twod_area", mcats.GetGeoJSONHandler(ctrl))
	// e.PATCH("/mcat/refresh_mesh_line", mcats.GeoRefreshHandler(project, ctrl))
	// e.PATCH("/mcat/refresh_break_line", mcats.GeoRefreshHandler(project, ctrl))
	// e.PATCH("/mcat/refresh_twod_area", mcats.GeoRefreshHandler(project, ctrl))
	// e.GET("/mcat/ras_index", mcatRasHandler.Index(ctrl.FS))
	// e.GET("/mcat/hms_index", mcatHmsHandler.Index(ctrl.FS))

	// Start server
	go func() {
		if err := e.Start(":5000"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down the server: %s", err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	e.Logger.Info("gracefully shutting down the server")

	if err := ctrl.Close(); err != nil {
		e.Logger.Errorf("error closing the controller: %x", err.Error())
	}
	e.Logger.Info("closed the controller and disconnected from the database")

	// shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	if ctrl.Bh.Config.AuthLevel > 0 {
		if err := ctrl.Bh.DB.Close(); err != nil {
			log.Error(err)
		} else {
			log.Info("closed connection to database")
		}
	}
}
