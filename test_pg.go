package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 测试不同的连接方式
	testConfigs := []struct {
		name string
		dsn  string
	}{
		{
			name: "disable SSL",
			dsn:  "host=pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com port=5432 user=wws password=Wws5201314 dbname=notice sslmode=disable",
		},
		{
			name: "require SSL",
			dsn:  "host=pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com port=5432 user=wws password=Wws5201314 dbname=notice sslmode=require",
		},
		{
			name: "prefer SSL",
			dsn:  "host=pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com port=5432 user=wws password=Wws5201314 dbname=notice sslmode=prefer",
		},
	}

	for _, tc := range testConfigs {
		fmt.Printf("\n测试 %s...\n", tc.name)
		db, err := sql.Open("postgres", tc.dsn)
		if err != nil {
			log.Printf("打开连接失败: %v", err)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Ping失败: %v", err)
			db.Close()
			continue
		}

		fmt.Printf("✅ %s 连接成功!\n", tc.name)

		// 查询版本
		var version string
		err = db.QueryRow("SELECT version()").Scan(&version)
		if err != nil {
			log.Printf("查询版本失败: %v", err)
		} else {
			fmt.Printf("PostgreSQL版本: %s\n", version)
		}

		db.Close()
	}
}
