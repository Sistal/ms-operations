package main

import (
	"fmt"
	"log"

	"ms-operations/internal/database"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	database.Connect()

	var tables []string
	database.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tables)

	fmt.Println("Tables found:")
	for _, t := range tables {
		fmt.Println(t)
	}
}
