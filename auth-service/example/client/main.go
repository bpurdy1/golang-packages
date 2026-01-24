package main

import (
	"context"
	"fmt"
	"log"

	authservice "github.com/bpurdy1/auth-service"
)

func main() {
	// Create client (defaults to in-memory database)
	client, err := authservice.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Create a user
	fmt.Println("--- Creating User ---")
	user, err := client.Users.CreateUser(ctx, authservice.CreateUserInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
		Email:     "john@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %s %s (ID: %d, UUID: %s)\n", user.FirstName, user.LastName, user.ID, user.Uuid)

	// Set some metadata
	fmt.Println("\n--- Setting Metadata ---")

	client.Metadata.Set(ctx, authservice.SetMetadataInput{UserID: user.ID, Key: "role", Value: "admin"})
	client.Metadata.Set(ctx, authservice.SetMetadataInput{UserID: user.ID, Key: "theme", Value: "dark"})
	fmt.Println("Set role=admin, theme=dark")

	// Get user with metadata
	fmt.Println("\n--- Get User With Metadata ---")
	userWithMeta, err := client.GetUserWithMetadata(ctx, user.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s %s\n", userWithMeta.FirstName, userWithMeta.LastName)
	fmt.Printf("Metadata: %v\n", userWithMeta.Metadata)

	// Authenticate
	fmt.Println("\n--- Authentication ---")
	authUser, err := client.Users.Authenticate(ctx, "johndoe", "password123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Authenticated: %s\n", authUser.Username)

	// Create user with metadata in one call
	fmt.Println("\n--- Create User With Metadata ---")
	user2, err := client.CreateUserWithMetadata(ctx, authservice.CreateUserInput{
		FirstName: "Jane",
		LastName:  "Smith",
		Username:  "janesmith",
		Password:  "password456",
		Email:     "jane@example.com",
	}, map[string]string{
		"role":       "user",
		"department": "engineering",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %s %s with metadata: %v\n", user2.FirstName, user2.LastName, user2.Metadata)

	// List all users
	fmt.Println("\n--- List Users ---")
	users, _ := client.Users.ListUsers(ctx, 10, 0)
	for _, u := range users {
		fmt.Printf("  - %s %s (%s)\n", u.FirstName, u.LastName, u.Username)
	}

	fmt.Println("\nDone!")
}
