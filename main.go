package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

// Funzione per caricare il template da una cartella
func getTemplatePath(filename string) string  {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd + "/" + filename
}


// Struttura per rappresentare la lattina
type MonsterCan struct {
	Name string
	Flavor string
	Image string
	Barcode string
}

func main()  {
	// Lista di lattine di Monster
	cans := []MonsterCan{
		{"Monster Energy", "Original", "image.jpg", "123456789"},
		{"Monster Ultra", "Zero Sugar", "image.jpg", "123456789"},
		{"Monster Mango Loco", "Mango", "image.jpg", "123456789"},
	}

	// Gestire la route per la home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Carica il template HTML da file con percorso completo
		tmplPath := getTemplatePath("index.html")
		tmpl, err := template.ParseFiles(tmplPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//Esegui il template con i dati delle lattine
			err = tmpl.Execute(w, cans)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
	})

	// Avvia il server sulla porta 8080
	fmt.Println("Server in esecuzione su http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}