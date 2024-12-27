package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
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

// Struttura per passare i dati al template
type PageData struct {
    Cans          []MonsterCan
    SuccessMessage string
	ErrorMessage string
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
		var pageData PageData
		if err != nil {
			pageData.ErrorMessage = "Errore nel caricare i dati"
		} else {
			// Se non c'è errore, carica il template
			pageData.Cans = cans
			pageData.SuccessMessage = "L'operazione è stata completata con successo!"
		}
			// http.Error(w, "Errore nel caricare i dati", http.StatusInternalServerError)
			// return
		// }

		// Crea un oggetto PageData con il messaggio di successo
		// pageData := PageData{
			// Cans: cans,
			// SuccessMessage: "L'operazione è stata completata con successo",
		// }

		// Carica il template HTML da file con percorso completo
		// tmplPath := getTemplatePath("index.html")
		tmpl, err := template.ParseFiles("templates/index.html")
			if err != nil {
				pageData.ErrorMessage = "Errore nel caricare il template: " + err.Error()
				// http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
				// return
			}

			//Esegui il template con i dati delle lattine
			err = tmpl.Execute(w, pageData)
			if err != nil {
				pageData.ErrorMessage = "Errore nell'esecuzione del template: " + err.Error()
				// http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
			}
	}

	// Gestione della route per aggiungere una nuova lattina
	func addHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Estrai i dati dal form
			name := r.FormValue("name")
			flavor := r.FormValue("flavor")
			barcode := r.FormValue("barcode")
			image, _, err := r.FormFile("image")
			if err != nil && err.Error() != "http: no such file" {
				http.Error(w, "Errore nell'upload dell'immagine", http.StatusInternalServerError)
				return
			}

			// Salva l'immagine (se presente)
			var imagePath string
			if image != nil {
				// Salva l'immagine nella cartella static/images
				imagePath = "static/images/" + barcode + ".jpg"
				dst, err := os.Create(imagePath)
				if err != nil {
					http.Error(w, "Errore nel salvare l'immagine", http.StatusInternalServerError)
					return
				}
				defer dst.Close()
				_, err = io.Copy(dst, image)
				if err != nil {
					http.Error(w, "Errore nel copiare l'immagine", http.StatusInternalServerError)
					return
				}
			}

			// Carica lattine esistenti
			cans, err := loadCans()
			if err != nil {
				http.Error(w, "Errore nel caricare i dati", http.StatusInternalServerError)
				return
			}

			// Aggiungi nuova lattina
			newCan := MonsterCan{Name: name, Flavor: flavor, Image: imagePath, Barcode: barcode}
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