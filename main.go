package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

// Struttura per rappresentare la lattina
type MonsterCan struct {
	Name string
	Flavor string
	Image string
	Barcode string
}


func getTemplatePath(templateName string) string  {
	// Ottieni il percorso assoluto della directory corrente
	exePath, err := os.Executable()
	if err != nil {
		panic("Errore nell'ottenere il percorso dell'eseguibile: " + err.Error())
	}
	dirPath := filepath.Dir(exePath)
	return filepath.Join(dirPath, "templates", templateName)
}



// Carica lattine dal file JSON
func loadCans() ([]MonsterCan, error)  {
	file, err := os.Open("cans.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cans []MonsterCan
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cans)
	if err != nil {
		return nil, err
	}

	return cans, nil
}

// Salva lattine nel file JSON
func saveCans(cans []MonsterCan) error  {
	file, err := os.Create("cans.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cans)
	if err != nil {
		return err
	}

	return nil
}

// Gestione della rotta per la homepage
	func homeHandler(w http.ResponseWriter, r *http.Request) {
		cans, err := loadCans()
		if err != nil {
			http.Error(w, "Errore nel caricare i dati", http.StatusInternalServerError)
			return
		}


		// Carica il template HTML da file con percorso completo
		// tmplPath := getTemplatePath("index.html")
		tmpl, err := template.ParseFiles("templates/index.html")
			if err != nil {
				http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
				return
			}

			//Esegui il template con i dati delle lattine
			err = tmpl.Execute(w, cans)
			if err != nil {
				http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
			}
	}

	// Gestione della route per aggiungere una nuova lattina
	func addHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Estrai i dati dal form
			name := r.FormValue("name")
			flavor := r.FormValue("flavor")
			image := r.FormValue("image")
			barcode := r.FormValue("barcode")

			// Carica lattine esistenti
			cans, err := loadCans()
			if err != nil {
				http.Error(w, "Errore nel caricare i dati", http.StatusInternalServerError)
				return
			}

			// Aggiungi nuova lattina
			newCan := MonsterCan{Name: name, Flavor: flavor, Image: image, Barcode: barcode}
			cans = append(cans, newCan)

			// Salva di nuovo nel file JSON
			err = saveCans(cans)
			if err != nil {
				http.Error(w, "Errore nel salvare i dati", http.StatusInternalServerError)
				return
			}

			// Redirect alla pagina principale dopo l'aggiunta
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Carica il template per il form di aggiunta
		// tmplPath := getTemplatePath("add.html")
		tmpl, err := template.ParseFiles("templates/add.html")
		if err != nil {
			http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Esegui il template
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
		}
	}

	// Gestione della route per rimuovere una lattina
	func removeHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			barcode := r.FormValue("barcode")

			
			// Carica lattine esistenti
			cans, err := loadCans()
			if err != nil {
				http.Error(w, "Errore nel caricare i dati", http.StatusInternalServerError)
				return
			}
			
			// Rimuovi la lattina con il codice a barre specificato
			var newCans []MonsterCan
			found := false
			for _, can := range cans {
				if can.Barcode != barcode {
					newCans = append(newCans, can)
					} else {
						found = true
					}
				}
				
				if !found {
					http.Error(w, "Lattina non trovata con il codice a barre: "+barcode, http.StatusInternalServerError)
					return
				}
				
				// Salva di nuovo nel file JSON
				err = saveCans(newCans)
				if err != nil {
					http.Error(w, "Errore nel salvare i dati", http.StatusInternalServerError)
					return
				}
				
				// Redirect alla pagina principale dopo la rimozione
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			
			// Carica il template HTML da file con percorso completo
			// tmplPath := getTemplatePath("remove.html")
			tmpl, err := template.ParseFiles("templates/remove.html")
			if err != nil {
				http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
				return
			}
			
			err = tmpl.Execute(w, nil)
			if err != nil {
				http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
			}
		}
		

		func main() {
			// Gestisci le rotte
			http.HandleFunc("/", homeHandler)
			http.HandleFunc("/add", addHandler)
			http.HandleFunc("/remove", removeHandler)
			
			
		
		// Avvia il server sulla porta 8080
		fmt.Println("Server in esecuzione su http://localhost:8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("Errore nell'avvio del server:", err)
		}
}