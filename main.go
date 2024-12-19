package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// Struttura per rappresentare la lattina
type MonsterCan struct {
	Name string
	Flavor string
	Image string
}

func main()  {
	// Lista di lattine di Monster
	cans := []MonsterCan{
		{"Monster Energy", "Original", "image.jpg"},
		{"Monster Ultra", "Zero Sugar", "image.jpg"},
		{"Monster Mango Loco", "Mango", "image.jpg"},
	}

	// Gestire la route per la home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Carica il template HTML
		tmpl, err := template.New("index").Parse(`
		<!DOCTYPE html>
			<html lang="it">
			<head>
				<meta charset="UTF-8">
				<title>Collezione Lattine di Monster</title>
				<style>
					body {
						font-family: Arial, sans-serif;
					}
					.container {
						display: flex;
						flex-wrap: wrap;
					}
					.item {
						margin: 10px;
						width: 200px;
					}
					img {
						width: 100%;
						height: auto;
					}
				</style>
			</head>
			<body>
				<h1>Collezione Lattine di Monster</h1>
				<div class="container">
					{{range .}}
					<div class="item">
						<h2>{{.Name}}</h2>
						<p><strong>Gusto:</strong> {{.Flavor}}</p>
						<img src="{{.Image}}" alt="{{.Name}}">
					</div>
					{{end}}
				</div>
			</body>
			</html>
			`)

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