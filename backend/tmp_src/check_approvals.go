package main
import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := "postgres://tisen:tisen123@localhost:5432/hrms?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var total int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM employees WHERE deleted_at IS NULL").Scan(&total)
	
	var withApproval int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM employees WHERE approval_line_id IS NOT NULL AND deleted_at IS NULL").Scan(&withApproval)
	
	fmt.Printf("Total active employees: %d\n", total)
	fmt.Printf("Employees with Approval Line mapped: %d\n", withApproval)

	rows, _ := pool.Query(ctx, "SELECT e.full_name, a.full_name FROM employees e JOIN employees a ON e.approval_line_id = a.id WHERE e.deleted_at IS NULL LIMIT 5")
	fmt.Println("\nSample mapped Approval Lines:")
	for rows.Next() {
		var eName, aName string
		rows.Scan(&eName, &aName)
		fmt.Printf("- %s -> %s\n", eName, aName)
	}
}
