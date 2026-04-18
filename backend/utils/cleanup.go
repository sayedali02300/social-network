package utils

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func CleanUpOrphanedImages(db *sql.DB, uploadDir string, table string){
	log.Println("Garbage Collector is working..")

	files, err := os.ReadDir(uploadDir)
	if err != nil{
		if os.IsNotExist(err){
			log.Println("Folder doesn't exist yet")
			return
		}
		log.Printf("Error reading folder %s: %v\n", uploadDir, err)
		return
	}

	query := `SELECT image_path FROM ` + table + ` WHERE image_path IS NOT NULL AND image_path != ''`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("DB error occured while cleaning: %v\n", err)
		return
	}
	defer rows.Close()

	validImages := make(map[string]bool)
	for rows.Next() {
		var dbPath string
		if err := rows.Scan(&dbPath); err != nil {
			continue
		}

		fileName := filepath.Base(dbPath)
		validImages[fileName] = true
	}

	deletedCount := 0
	for _, file := range files {
		if file.IsDir(){
			continue
		}

		fileName := file.Name()

		if !validImages[fileName] {
			fullPath := filepath.Join(uploadDir, fileName)
			err := os.Remove(fullPath)
			if err != nil {
				log.Printf("failed to delete the orphan img %s: %v\n", fullPath, err)
			} else{
				log.Printf("extra image was cleaned: %s\n", fileName)
				deletedCount++
			}
		}
	}
	log.Printf("%d Images were deleted.", deletedCount)
}