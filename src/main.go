package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("          SERVER VOXELIZATION AKTIF               ")
	fmt.Println("==================================================")
	fmt.Println("Menunggu input file...")
	fmt.Println("Silakan buka browser dan akses: http://localhost:8080")

	// Mendaftarkan endpoint URL
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/voxelize", handleVoxelize)

	port := ":8080"
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server Error: %v\n", err)
	}
}
