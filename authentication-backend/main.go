package main

import (
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/authentication-backend/internal"
	_ "github.com/glimesh/broadcast-box/authentication-backend/migrations"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	app.OnBeforeServe().Add(internal.HandleServeEvent)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
