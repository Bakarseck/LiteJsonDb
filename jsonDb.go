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

// Ajoute une contrainte dans la base de données
func (db *JsonDB) SetConstraint(category, constraintType, field string) {
	constraints, ok := db.DB["constraints"].(map[string]interface{})
	if !ok {
		constraints = make(map[string]interface{})
		db.DB["constraints"] = constraints
	}

	categoryConstraints, ok := constraints[category].(map[string]interface{})
	if !ok {
		categoryConstraints = make(map[string]interface{})
		constraints[category] = categoryConstraints
	}

	fields, ok := categoryConstraints[constraintType].([]interface{})
	if !ok {
		fields = []interface{}{}
	}

	// Ajouter le champ s'il n'est pas déjà présent
	for _, f := range fields {
		if f == field {
			return // Le champ est déjà présent pour cette contrainte
		}
	}

	categoryConstraints[constraintType] = append(fields, field)
	db.saveDB()
}

// Vérifie si un champ est soumis à une contrainte spécifique pour une catégorie
func (db *JsonDB) IsFieldConstrained(category, constraintType, field string) bool {
	constraints, ok := db.DB["constraints"].(map[string]interface{})
	if !ok {
		return false
	}

	categoryConstraints, ok := constraints[category].(map[string]interface{})
	if !ok {
		return false
	}

	fields, ok := categoryConstraints[constraintType].([]interface{})
	if !ok {
		return false
	}

	for _, f := range fields {
		if f == field {
			return true
		}
	}
	return false
}

// Vérifie si une valeur est unique pour un champ soumis à une contrainte unique
func (db *JsonDB) IsUnique(category, field, value string) bool {
	if !db.IsFieldConstrained(category, "unique", field) {
		return true // Si le champ n'est pas contraint à l'unicité, on ne vérifie pas
	}

	for key, record := range db.DB {
		if key[:len(category)] == category {
			recordMap, ok := record.(map[string]interface{})
			if ok && recordMap[field] == value {
				return false
			}
		}
	}
	return true
}

// SetDataWithAutoIncrement ajoute une entrée avec un ID auto-incrémenté
func (db *JsonDB) SetDataWithAutoIncrement(category string, value map[string]interface{}) (int, error) {
	// Vérifier l'unicité pour tous les champs soumis à une contrainte unique
	for field := range value {
		if !db.IsUnique(category, field, value[field].(string)) {
			return 0, fmt.Errorf("%s %s already exists", field, value[field].(string))
		}
	}

	nextID := db.GetNextID(category)
	key := fmt.Sprintf("%s_%d", category, nextID)
	db.SetData(key, value)
	return nextID, nil
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

func main() {
	db := NewJsonDB(DB_FILE)

	// Définir le champ "username" comme unique pour la catégorie "user"
	db.SetConstraint("user", "unique", "username")

	// Ajouter une entrée dans la catégorie "user" avec auto-incrément
	userData := map[string]interface{}{
		"username": "Alice",
		"password": HashPassword("password123"),
	}

	userID, err := db.SetDataWithAutoIncrement("user", userData)
	if err != nil {
		fmt.Printf("Error adding user: %v\n", err)
	} else {
		fmt.Printf("User added with ID: %d\n", userID)
	}

	// Tenter d'ajouter le même utilisateur
	userID, err = db.SetDataWithAutoIncrement("user", userData)
	if err != nil {
		fmt.Printf("Error adding user: %v\n", err)
	} else {
		fmt.Printf("User added with ID: %d\n", userID)
	}
}
