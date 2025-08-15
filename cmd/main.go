package main

import (
	"log"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/localhandlers"
	"github.com/DanilNaum/secret-storage-client/internal/models"
	"github.com/DanilNaum/secret-storage-client/internal/serverhandlers"
	"github.com/DanilNaum/secret-storage-client/internal/storage"
	"github.com/DanilNaum/secret-storage-client/internal/ui"
	"github.com/DanilNaum/secret-storage-client/pkg/cache"
)

func main() {
	stor := storage.NewStorage()

	serverCache := cache.NewCache[models.Record](constants.DefaultCacheTTL)
	
	localHandlers := localhandlers.NewLocalHandlers(stor)
	
	serverURL := "https://api.secretstorage.com"
	
	serverHandlers := serverhandlers.NewServerHandlers(stor, serverCache, serverURL)
	
	app := ui.NewApp(stor, localHandlers, serverHandlers)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}