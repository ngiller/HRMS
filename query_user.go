package main
import (
	"context"
	"fmt"
	"hrms-backend/internal/database"
)
func main() {
	database.Connect()
	var email string
	err := database.Pool.QueryRow(context.Background(), "SELECT email FROM employees WHERE deleted_at IS NULL LIMIT 1").Scan(&email)
	if err != nil { fmt.Println("err:", err); return }
	fmt.Println(email)
}
