package main

import (
	"context"
	"fmt"
	"github.com/NellinLin/tp_db_forum/cmd/api/handlers"
	"github.com/NellinLin/tp_db_forum/internal/forum"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	connectionString = "postgres://forum:1111@localhost:5432/forum?sslmode=disable"
	//connectionString = "postgres://postgres:1111@localhost:5432/forum?sslmode=disable"
	host             = "0.0.0.0:5000"
	maxConn          = 1000
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	config, err := pgx.ParseURI(connectionString)
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: maxConn,
		})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = LoadSchemaSQL(db)
	if err != nil {
		fmt.Println(err)
	}

	forumService := forum.NewForumService(db)
	postService := forum.NewPostService(db)
	threadService := forum.NewThreadService(db)
	userService := forum.NewUserService(db)
	serviceService := forum.NewServiceService(db)

	serverErrors := make(chan error, 1)

	api := &http.Server{
		Addr:    host,
		Handler: handlers.API(&log, forumService, postService, threadService, userService, serviceService),
	}

	go func() {
		log.Info().Msgf("Start API Listening %s", host)
		serverErrors <- api.ListenAndServe()
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error().Msgf("Error starting server: %v", err)

	case <-osSignals:
		log.Info().Msg("Start shutdown...")
		if err := api.Shutdown(context.Background()); err != nil {
			log.Error().Msgf("Graceful shutdown error: %v", err)
			os.Exit(1)
		}
	}
}

const dbSchema = "db.sql"

func LoadSchemaSQL(db *pgx.ConnPool) error {

	content, err := ioutil.ReadFile(dbSchema)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(content)); err != nil {
		return err
	}
	tx.Commit()
	return nil
}
