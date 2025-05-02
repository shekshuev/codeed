package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a MongoDB document in the "users" collection.
type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`     // MongoDB ObjectID (auto-generated if omitted)
	TelegramUsername string             `bson:"telegram_username"` // Telegram username (external identity)
	Username         string             `bson:"username"`          // Unique username (e.g., Telegram handle)
	FirstName        string             `bson:"first_name"`        // User's first name
	LastName         string             `bson:"last_name"`         // User's last name
	Role             string             `bson:"role"`              // "admin" or "student"
	CreatedAt        time.Time          `bson:"created_at"`        // Timestamp when user was created
	UpdatedAt        time.Time          `bson:"updated_at"`        // Timestamp when user was last updated
	DeletedAt        *time.Time         `bson:"deleted_at"`        // Optional timestamp when user was deleted (soft delete)
}

// CreateUserDTO is used when creating a new user from client data (e.g. after Telegram login).
type CreateUserDTO struct {
	TelegramUsername string `json:"telegram_username"` // Telegram user ID
	Username         string `json:"username"`          // Username to register
	FirstName        string `json:"first_name"`        // First name of user
	LastName         string `json:"last_name"`         // Last name of user
	Role             string `json:"role"`              // User role: "admin" or "student"
}

// UpdateUserDTO is used to update an existing user. Fields are optional.
type UpdateUserDTO struct {
	Username  *string `json:"username,omitempty"`   // Optional new username
	FirstName *string `json:"first_name,omitempty"` // Optional new first name
	LastName  *string `json:"last_name,omitempty"`  // Optional new last name
}

// ReadUserDTO is a presentation-layer representation of the user,
// used for returning data to the frontend via JSON.
type ReadUserDTO struct {
	ID               string `json:"id"`                // Mongo ObjectID as hex string
	TelegramUsername string `json:"telegram_username"` // Telegram user ID
	Username         string `json:"username"`          // Username
	FirstName        string `json:"first_name"`        // First name
	LastName         string `json:"last_name"`         // Last name
	Role             string `json:"role"`              // "admin" or "student"
	CreatedAt        string `json:"created_at"`        // ISO 8601 timestamp (RFC3339)
	UpdatedAt        string `json:"updated_at"`        // ISO 8601 timestamp (RFC3339)
}

// ToReadUserDTO converts the User entity into a ReadUserDTO suitable for JSON output.
func (u User) ToReadUserDTO() *ReadUserDTO {
	return &ReadUserDTO{
		ID:               u.ID.Hex(),
		TelegramUsername: u.TelegramUsername,
		Username:         u.Username,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Role:             u.Role,
		CreatedAt:        u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        u.UpdatedAt.Format(time.RFC3339),
	}
}

// ToUserFromCreateDTO maps the incoming CreateUserDTO into a full User model,
// generating ObjectID and setting CreatedAt.
func (dto CreateUserDTO) ToUserFromCreateDTO() User {
	return User{
		ID:               primitive.NewObjectID(),
		TelegramUsername: dto.TelegramUsername,
		Username:         dto.Username,
		FirstName:        dto.FirstName,
		LastName:         dto.LastName,
		Role:             dto.Role,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

// ToBsonUpdateFromUpdateDTO creates a BSON update map (e.g. {"$set": ...}) from UpdateUserDTO.
// Only non-nil fields are included. Returns empty map if nothing to update.
func (dto UpdateUserDTO) ToBsonUpdateFromUpdateDTO() bson.M {
	update := bson.M{}
	if dto.Username != nil {
		update["username"] = *dto.Username
	}
	if dto.FirstName != nil {
		update["first_name"] = *dto.FirstName
	}
	if dto.LastName != nil {
		update["last_name"] = *dto.LastName
	}
	if len(update) == 0 {
		return bson.M{}
	}
	update["updated_at"] = time.Now().UTC()
	return bson.M{"$set": update}
}
