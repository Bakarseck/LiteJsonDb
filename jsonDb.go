package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const DATABASE_DIR = "database"
const DB_FILE = "database.json"

type JsonDB struct {
	FilePath string
	DB       map[string]interface{}
}

func NewJsonDB(filename string) *JsonDB {
	db := &JsonDB{
		FilePath: filepath.Join(DATABASE_DIR, filename),
		DB:       make(map[string]interface{}),
	}
	db.initDB()
	return db
}

func (db *JsonDB) initDB() {
	if _, err := os.Stat(DATABASE_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(DATABASE_DIR, os.ModePerm)
		if err != nil {
			log.Fatalf("Oops! Unable to create the database directory. Error details: %v", err)
		}
	}

	if _, err := os.Stat(db.FilePath); os.IsNotExist(err) {
		err := os.WriteFile(db.FilePath, []byte("{}"), 0644)
		if err != nil {
			log.Fatalf("Oops! Unable to create the database file. Error details: %v", err)
		}
	} else {
		db.loadDB()
	}
}

func (db *JsonDB) loadDB() {
	fileData, err := os.ReadFile(db.FilePath)
	if err != nil {
		log.Fatalf("Oops! Unable to load the database file. Error details: %v", err)
	}

	err = json.Unmarshal(fileData, &db.DB)
	if err != nil {
		log.Fatalf("Oops! The database file is not valid JSON. Error details: %v", err)
	}
}

func (db *JsonDB) saveDB() {
	fileData, err := json.MarshalIndent(db.DB, "", "  ")
	if err != nil {
		log.Fatalf("Oops! Unable to save the database file. Error details: %v", err)
	}

	err = os.WriteFile(db.FilePath, fileData, 0644)
	if err != nil {
		log.Fatalf("Oops! Unable to save the database file. Error details: %v", err)
	}
}

func (db *JsonDB) SetData(key string, value interface{}) {
	db.DB[key] = value
	db.saveDB()
}

func (db *JsonDB) GetData(key string) interface{} {
	if value, exists := db.DB[key]; exists {
		return value
	}
	log.Printf("Oops! The key '%s' does not exist.", key)
	return nil
}

func (db *JsonDB) DeleteData(key string) {
	if _, exists := db.DB[key]; exists {
		delete(db.DB, key)
		db.saveDB()
	} else {
		log.Printf("Oops! The key '%s' does not exist.", key)
	}
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func CheckPassword(storedHash, password string) bool {
	return storedHash == HashPassword(password)
}
