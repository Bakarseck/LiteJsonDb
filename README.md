# JsonDB - Simple JSON Database in Go

`JsonDB` is a lightweight, file-based database management system implemented in Go. It provides an easy way to store, retrieve, and manage data in a JSON file format. This project is particularly useful for small applications that do not require the overhead of a full-fledged database management system.

## Features

- **File-based storage**: All data is stored in a single JSON file.
- **Auto-incremented IDs**: Automatically generate unique IDs for records.
- **Category-based organization**: Data is organized into categories, each with its own set of records.
- **Hashing passwords**: Provides built-in functionality to hash and verify passwords.
- **Simple and extensible**: The code is straightforward, making it easy to extend and customize.

## Getting Started

### Prerequisites

- Go 1.18 or higher installed on your machine.
- Basic understanding of Go programming.

### Installation

1. Clone the repository to your local machine:

   ```bash
   git clone https://github.com/yourusername/jsondb-go.git
   cd jsondb-go
   ```

2. Build the project:

   ```bash
   go build -o jsondb
   ```

3. Run the project:

   ```bash
   go run main.go
   ```

### Project Structure

- **`main.go`**: The main entry point of the application. Demonstrates how to use the `JsonDB` to store and retrieve data.
- **`jsondb.go`**: The core library containing all the logic for interacting with the JSON database.
- **`database/`**: Directory where the JSON database file (`database.json`) is stored.

### Usage

#### Initializing the Database

The `JsonDB` is initialized by specifying the filename where the data will be stored:

```go
db := NewJsonDB("database.json")
```

#### Adding Data with Auto-Incremented IDs

You can add data to a category with an auto-incremented ID:

```go
userID := db.SetDataWithAutoIncrement("user", map[string]interface{}{
    "username": "Alice",
    "password": HashPassword("password123"),
})
fmt.Printf("User added with ID: %d\n", userID)
```

This will store the user data under `user/1` if it’s the first entry in the `user` category.

#### Retrieving Data

You can retrieve data by specifying the category and ID:

```go
user := db.GetData(fmt.Sprintf("user/%d", userID))
fmt.Printf("User %d: %v\n", userID, user)
```

This retrieves the user data stored under `user/1`.

#### Deleting Data

To delete a record, specify the category and ID:

```go
db.DeleteData(fmt.Sprintf("user/%d", userID))
```

This deletes the user data stored under `user/1`.

### Hashing Passwords

`JsonDB` includes utility functions for hashing and verifying passwords:

```go
hashedPassword := HashPassword("mysecretpassword")
fmt.Println("Hashed Password:", hashedPassword)

isValid := CheckPassword(hashedPassword, "mysecretpassword")
fmt.Println("Password is valid:", isValid)
```

### JSON Structure

The JSON database structure will look something like this:

```json
{
  "auto_increment": {
    "user": 1
  },
  "user": {
    "1": {
      "username": "Alice",
      "password": "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f"
    }
  }
}
```

### License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
