package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bpurdy1/auth-service/account"
	"github.com/bpurdy1/auth-service/config"
)

func main() {
	// Use in-memory for example (set DB_PATH env var to use file)
	if os.Getenv("DB_PATH") == "" {
		os.Setenv("DB_PATH", ":memory:")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("Using database: %s\n", cfg.DBPath)

	// Open database connection
	db, err := cfg.OpenDB()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := account.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	fmt.Println("Database migrated successfully")

	// Create user service
	svc := account.NewUserService(db)
	ctx := context.Background()

	// Create a user
	fmt.Println("\n--- Creating User ---")
	user, err := svc.CreateUser(ctx, account.CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
		Email:     "john.doe@example.com",
	})
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	fmt.Printf("Created user: ID=%d, UUID=%s, Username=%s\n", user.ID, user.Uuid, user.Username)

	// Authenticate
	fmt.Println("\n--- Authenticating ---")
	authUser, err := svc.Authenticate(ctx, "johndoe", "securepassword123")
	if err != nil {
		log.Fatalf("authentication failed: %v", err)
	}
	fmt.Printf("Authenticated: %s %s (%s)\n", authUser.FirstName, authUser.LastName, authUser.Email)

	// Wrong password
	_, err = svc.Authenticate(ctx, "johndoe", "wrongpassword")
	if err != nil {
		fmt.Printf("Wrong password rejected: %v\n", err)
	}

	// Get user by various methods
	fmt.Println("\n--- Fetching User ---")
	byID, _ := svc.GetUserByID(ctx, user.ID)
	fmt.Printf("By ID:       %s %s\n", byID.FirstName, byID.LastName)

	byUUID, _ := svc.GetUserByUUID(ctx, user.Uuid)
	fmt.Printf("By UUID:     %s %s\n", byUUID.FirstName, byUUID.LastName)

	byUsername, _ := svc.GetUserByUsername(ctx, "johndoe")
	fmt.Printf("By Username: %s %s\n", byUsername.FirstName, byUsername.LastName)

	byEmail, _ := svc.GetUserByEmail(ctx, "john.doe@example.com")
	fmt.Printf("By Email:    %s %s\n", byEmail.FirstName, byEmail.LastName)

	// Update user
	fmt.Println("\n--- Updating User ---")
	updated, err := svc.UpdateUser(ctx, user.ID, account.UpdateUserInput{
		FirstName: "Jonathan",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "jonathan.doe@example.com",
	})
	if err != nil {
		log.Fatalf("failed to update user: %v", err)
	}
	fmt.Printf("Updated: %s %s (%s)\n", updated.FirstName, updated.LastName, updated.Email)

	// Update password
	fmt.Println("\n--- Updating Password ---")
	_, err = svc.UpdatePassword(ctx, user.ID, "newpassword456")
	if err != nil {
		log.Fatalf("failed to update password: %v", err)
	}
	fmt.Println("Password updated")

	// Verify new password works
	_, err = svc.Authenticate(ctx, "johndoe", "newpassword456")
	if err != nil {
		log.Fatalf("new password authentication failed: %v", err)
	}
	fmt.Println("New password authentication successful")

	// Create more users for listing
	fmt.Println("\n--- Creating More Users ---")
	for i := 1; i <= 3; i++ {
		_, err := svc.CreateUser(ctx, account.CreateUserInput{
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  "Test",
			Username:  fmt.Sprintf("user%d", i),
			Password:  "password123",
			Email:     fmt.Sprintf("user%d@example.com", i),
		})
		if err != nil {
			log.Fatalf("failed to create user%d: %v", i, err)
		}
		fmt.Printf("Created user%d\n", i)
	}

	// List users
	fmt.Println("\n--- Listing Users ---")
	users, err := svc.ListUsers(ctx, 10, 0)
	if err != nil {
		log.Fatalf("failed to list users: %v", err)
	}
	for _, u := range users {
		fmt.Printf("  - %s %s (%s) - Created: %s\n", u.FirstName, u.LastName, u.Username, u.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Count users
	count, _ := svc.CountUsers(ctx)
	fmt.Printf("\nTotal users: %d\n", count)

	// Delete user
	fmt.Println("\n--- Deleting User ---")
	err = svc.DeleteUser(ctx, user.ID)
	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}
	fmt.Printf("Deleted user ID=%d\n", user.ID)

	count, _ = svc.CountUsers(ctx)
	fmt.Printf("Remaining users: %d\n", count)
}
