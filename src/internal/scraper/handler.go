package scraper

import (
	"context"
	"encoding/json"
	"log"
	"mimir-scrapper/src/internal/scraper/fetcher"
	"mimir-scrapper/src/internal/scraper/parser"
	"mimir-scrapper/src/pkg/repository"
	"mimir-scrapper/src/pkg/utils"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ScrapeHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, pool *pgxpool.Pool) {
	const (
		url       = "https://www.senat.cz/xqw/xervlet/pssenat/finddoc?typdok=steno"
		outputDir = "data/raw_data" // Directory for storing documents
	)

	// Ensure the output directory exists
	if err := utils.EnsureDir(outputDir); err != nil {
		log.Println("Error creating output directory:", err)
		http.Error(w, "Failed to set up storage", http.StatusInternalServerError)
		return
	}

	// Fetch HTML documents
	documents, err := fetcher.FetchPage(url)
	if err != nil {
		log.Println("Error fetching documents:", err)
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	// Process and save documents
	parsedDocuments, err := processAndSaveDocuments(ctx, pool, documents)
	if err != nil {
		log.Println("Error processing documents:", err)
		http.Error(w, "Failed to process documents", http.StatusInternalServerError)
		return
	}

	// Respond with the number of successfully processed documents
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]int{"processed_documents": len(parsedDocuments)})
}

func processAndSaveDocuments(ctx context.Context, pool *pgxpool.Pool, documents []string) ([]interface{}, error) {
	institutionName := "Senát"
	occasionName := "meeting"
	var parsedDocuments []interface{}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	repo := repository.New(tx)

	// Find or create institution
	institution, err := repo.FindInstitutionByName(ctx, institutionName)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("Institution '%s' not found, creating it.", institutionName)
			institution, err = repo.InsertInstitution(ctx, institutionName)
			if err != nil {
				log.Printf("Error creating institution: %v", err)
				return nil, err
			}
		} else {
			log.Printf("Error finding institution: %v", err)
			return nil, err
		}
	}

	// Find or create occasion
	occasion, err := repo.FindOccasionByName(ctx, occasionName)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("Occasion '%s' not found, creating it.", occasionName)
			occasion, err = repo.InsertOccasion(ctx, occasionName)
			if err != nil {
				log.Printf("Error creating occasion: %v", err)
				return nil, err
			}
		} else {
			log.Printf("Error finding occasion: %v", err)
			return nil, err
		}
	}

	// Insert session
	session, err := repo.InsertSession(ctx, repository.InsertSessionParams{
		InstitutionID: institution.ID,
		OccasionID:    occasion.ID,
		DateTime:      time.Now(),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
		return nil, err
	}
	log.Printf("Session created with ID: %s", session.ID.String())

	// Process documents
	for index, content := range documents {
		// Parse the HTML document
		transcript, err := parser.ParseHTMLDocument(content)
		if err != nil {
			log.Printf("Error parsing document %d: %v", index, err)
			continue
		}

		parsedDocuments = append(parsedDocuments, transcript)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	return parsedDocuments, nil
}
