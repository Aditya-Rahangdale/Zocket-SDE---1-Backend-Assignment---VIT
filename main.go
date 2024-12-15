
// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// Database connection variable
var db *sql.DB

func init() {
	// Load environment variables
	connStr := os.Getenv("POSTGRES_CONN")

	// Connect to the database
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	log.Println("Database connection established.")
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	type Product struct {
		UserID            int      `json:"user_id"`
		ProductName       string   `json:"product_name"`
		ProductDescription string   `json:"product_description"`
		ProductImages     []string `json:"product_images"`
		ProductPrice      float64  `json:"product_price"`
	}

	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO products (user_id, product_name, product_description, product_images, product_price)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	var productID int
	err := db.QueryRow(query, product.UserID, product.ProductName, product.ProductDescription, pq.Array(product.ProductImages), product.ProductPrice).Scan(&productID)
	if err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		log.Printf("Error inserting product: %v", err)
		return
	}

	response := map[string]interface{}{
		"message": "Product created successfully",
		"product_id": productID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func GetProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID := params["id"]

	query := `
		SELECT user_id, product_name, product_description, product_images, compressed_product_images, product_price
		FROM products WHERE id = $1
	`

	var (
		userID               int
		productName          string
		productDescription   string
		productImages        []string
		compressedImages     []string
		productPrice         float64
	)

	row := db.QueryRow(query, productID)
	err := row.Scan(&userID, &productName, &productDescription, pq.Array(&productImages), pq.Array(&compressedImages), &productPrice)
	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve product", http.StatusInternalServerError)
		log.Printf("Error retrieving product: %v", err)
		return
	}

	response := map[string]interface{}{
		"user_id":               userID,
		"product_name":          productName,
		"product_description":   productDescription,
		"product_images":        productImages,
		"compressed_images":     compressedImages,
		"product_price":         productPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	priceMin := r.URL.Query().Get("price_min")
	priceMax := r.URL.Query().Get("price_max")
	productName := r.URL.Query().Get("product_name")

	query := "SELECT id, user_id, product_name, product_description, product_images, product_price FROM products WHERE user_id = $1"
	args := []interface{}{userID}

	if priceMin != "" {
		query += " AND product_price >= $2"
		priceMinVal, _ := strconv.ParseFloat(priceMin, 64)
		args = append(args, priceMinVal)
	}
	if priceMax != "" {
		query += " AND product_price <= $3"
		priceMaxVal, _ := strconv.ParseFloat(priceMax, 64)
		args = append(args, priceMaxVal)
	}
	if productName != "" {
		query += " AND product_name ILIKE $4"
		args = append(args, "%"+productName+"%")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		log.Printf("Error retrieving products: %v", err)
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var (
			id                 int
			userID            int
			productName       string
			productDescription string
			productImages     []string
			productPrice      float64
		)
		if err := rows.Scan(&id, &userID, &productName, &productDescription, pq.Array(&productImages), &productPrice); err != nil {
			http.Error(w, "Failed to parse product data", http.StatusInternalServerError)
			log.Printf("Error scanning product row: %v", err)
			return
		}
		products = append(products, map[string]interface{}{
			"id":                  id,
			"user_id":             userID,
			"product_name":        productName,
			"product_description": productDescription,
			"product_images":      productImages,
			"product_price":       productPrice,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func main() {
	r := mux.NewRouter()

	// Register API endpoints
	r.HandleFunc("/products", CreateProductHandler).Methods("POST")
	r.HandleFunc("/products/{id}", GetProductByIDHandler).Methods("GET")
	r.HandleFunc("/products", GetProductsHandler).Methods("GET")

	// Start the server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
