package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/services/member/migrations"
	"github.com/zeromicro/go-zero/core/conf"
	"log"
	"time"
)

func main() {
	file := flag.String("f", "services/member/rpc/etc/member.yaml", "configuration file")
	flag.Parse()
	if flag.NArg() != 1 || flag.Arg(0) != "migrate" {
		log.Fatal("usage: member [-f config] migrate")
	}
	var c struct{ MySQL dbx.MySQLConfig }
	conf.MustLoad(*file, &c, conf.UseEnv())
	db, err := dbx.OpenMySQL(c.MySQL)
	if err != nil {
		log.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err = migrations.Up(ctx, pool); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Member migrations applied.")
}
