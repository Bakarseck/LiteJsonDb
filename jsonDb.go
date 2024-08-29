package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		// Initialiser auto_increment correctement
		initialData := map[string]interface{}{
			"auto_increment": make(map[string]interface{}),
		}
		fileData, _ := json.MarshalIndent(initialData, "", "  ")
		err := os.WriteFile(db.FilePath, fileData, 0644)
		if err != nil {
			log.Fatalf("Oops! Unable to create the database file. Error details: %v", err)
		}
	} else {
		db.loadDB()
	}
}

// GetNextID retourne l'ID suivant pour une catégorie donnée et l'incrémente
func (db *JsonDB) GetNextID(category string) int {
	autoIncrement, ok := db.DB["auto_increment"].(map[string]interface{})
	if !ok || autoIncrement == nil {
		// Si auto_increment n'existe pas ou est nil, l'initialiser correctement
		autoIncrement = make(map[string]interface{})
		db.DB["auto_increment"] = autoIncrement
	}

	if _, exists := autoIncrement[category]; !exists {
		autoIncrement[category] = 1
		db.saveDB()
		return 1
	}
	nextID := int(autoIncrement[category].(float64)) + 1
	autoIncrement[category] = nextID
	db.saveDB()
	return nextID
}

// SetDataWithAutoIncrement ajoute une entrée avec un ID auto-incrémenté
func (db *JsonDB) SetDataWithAutoIncrement(category string, value interface{}) int {
	nextID := db.GetNextID(category)

	// Vérifier si la catégorie existe déjà dans la base de données
	if _, exists := db.DB[category]; !exists {
		db.DB[category] = make(map[string]interface{}) // Créer un nouveau map pour cette catégorie
	}

	// Ajouter les données sous la clé de catégorie avec le sous-ID
	categoryMap := db.DB[category].(map[string]interface{})
	categoryMap[fmt.Sprintf("%d", nextID)] = value

	// Sauvegarder les modifications
	db.saveDB()

	return nextID
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
	// Extraire la catégorie et l'ID
	category := key[:strings.Index(key, "/")]
	id := key[strings.Index(key, "/")+1:]

	// Vérifier si la catégorie et l'ID existent
	if categoryMap, exists := db.DB[category].(map[string]interface{}); exists {
		if value, exists := categoryMap[id]; exists {
			return value
		}
	}
	log.Printf("Oops! The key '%s' does not exist.", key)
	return nil
}

func (db *JsonDB) DeleteData(key string) {
	// Extraire la catégorie et l'ID
	category := key[:strings.Index(key, "/")]
	id := key[strings.Index(key, "/")+1:]

	// Vérifier si la catégorie et l'ID existent
	if categoryMap, exists := db.DB[category].(map[string]interface{}); exists {
		if _, exists := categoryMap[id]; exists {
			delete(categoryMap, id)
			db.saveDB()
		} else {
			log.Printf("Oops! The ID '%s' does not exist in category '%s'.", id, category)
		}
	} else {
		log.Printf("Oops! The category '%s' does not exist.", category)
	}
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func CheckPassword(storedHash, password string) bool {
	return storedHash == HashPassword(password)
}
