package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/timezone"
	"log"
	"os"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/migrations"
	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	if err := timezone.Configure(); err != nil {
		log.Fatal(err)
	}
	configFile := flag.String("f", "services/system/rpc/etc/system.yaml", "configuration file")
	username := flag.String("username", "admin", "initial administrator username")
	displayName := flag.String("name", "管理员", "initial display name")
	flag.Parse()
	if flag.NArg() != 1 || (flag.Arg(0) != "migrate" && flag.Arg(0) != "bootstrap" && flag.Arg(0) != "disable" && flag.Arg(0) != "sync-resources") {
		log.Fatal("usage: system [-f config] [-username admin] [-name display] migrate|bootstrap|disable|sync-resources")
	}
	var c struct{ MySQL dbx.MySQLConfig }
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	open := dbx.OpenMySQL
	if flag.Arg(0) == "migrate" {
		open = dbx.OpenMigrationMySQL
	}
	db, err := open(c.MySQL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if flag.Arg(0) == "migrate" {
		if err = migrations.Up(ctx, pool); err != nil {
			log.Fatal(err)
		}
		fmt.Println("System migrations applied.")
		return
	}
	manager, err := management.New(repository.NewRBAC(db), authorization.NewDeferred(), nil, time.Local)
	if err != nil {
		log.Fatal(err)
	}
	if flag.Arg(0) == "sync-resources" {
		if err = manager.SyncResources(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Permission catalog synchronized; existing grants preserved.")
		return
	}
	service := auth.NewCommands(repository.NewAuth(db))
	if flag.Arg(0) == "disable" {
		if err = service.Disable(ctx, *username); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Administrator %s disabled; sessions revoked.\n", *username)
		return
	}
	created, err := service.Bootstrap(ctx, *username, os.Getenv("ADMIN_BOOTSTRAP_PASSWORD"), *displayName)
	if err != nil {
		log.Fatal("bootstrap failed: check database and username/password requirements (12–72 password bytes)")
	}
	if err = manager.SyncResources(ctx); err != nil {
		log.Fatal(err)
	}
	initialized, err := manager.BootstrapRoot(ctx, *username)
	if err != nil {
		log.Fatal(err)
	}
	if initialized {
		fmt.Println("Initial super administrator role assigned.")
	}
	if created {
		fmt.Printf("Administrator %s created.\n", *username)
	} else {
		fmt.Printf("Administrator %s already exists; no changes made.\n", *username)
	}
}
