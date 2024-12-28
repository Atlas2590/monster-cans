package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Struttura per rappresentare la lattina
type MonsterCan struct {
    Name    string
    Flavor  string
    Image   string
    Barcode string
}

// Struttura per passare i dati al template
type PageData struct {
    Cans           []MonsterCan
    SuccessMessage string
    ErrorMessage   string
}

// Connessione al database SQLite
func openDB() (*sql.DB, error) {
    // Crea o apri il database SQLite
    db, err := sql.Open("sqlite3", "./cans.db")
    if err != nil {
        return nil, err
    }

    // Crea la tabella "cans" se non esiste
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS cans (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        flavor TEXT,
        image TEXT,
        barcode TEXT UNIQUE
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, err
    }
    return db, nil
}

// Carica lattine dal database
func loadCans() ([]MonsterCan, error) {
    db, err := openDB()
    if err != nil {
        return nil, err
    }
    defer db.Close()

    rows, err := db.Query("SELECT name, flavor, image, barcode FROM cans")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cans []MonsterCan
    for rows.Next() {
        var can MonsterCan
        if err := rows.Scan(&can.Name, &can.Flavor, &can.Image, &can.Barcode); err != nil {
            return nil, err
        }
        cans = append(cans, can)
    }

    return cans, nil
}

// Aggiungi una lattina al database
func saveCans(can MonsterCan) error {
    db, err := openDB()
    if err != nil {
        return err
    }
    defer db.Close()

    // Inserisci una nuova lattina
    _, err = db.Exec("INSERT INTO cans (name, flavor, image, barcode) VALUES (?, ?, ?, ?)", can.Name, can.Flavor, can.Image, can.Barcode)
    if err != nil {
        return err
    }

    return nil
}

// Rimuovi una lattina dal database
func removeCan(barcode string) error {
    db, err := openDB()
    if err != nil {
        return err
    }
    defer db.Close()

    _, err = db.Exec("DELETE FROM cans WHERE barcode = ?", barcode)
    return err
}

// Gestione della rotta per la homepage
func homeHandler(w http.ResponseWriter, r *http.Request) {
    var pageData PageData

    // Rimuovi lattina se il form è stato inviato
    if r.Method == http.MethodPost {
        if barcode := r.FormValue("barcode"); barcode != "" {
            err := removeCan(barcode)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }
    }
    // Carica lattine dal database
    cans, err := loadCans()
    if err != nil {
        pageData.ErrorMessage = "Errore nel caricare i dati: " + err.Error()
    } else {
        pageData.Cans = cans
    }
    // Carica il template HTML
    tmpl, err := template.ParseFiles("templates/index.html")
    if err != nil {
        pageData.ErrorMessage = "Errore nel caricare il template: " + err.Error()
    }

    // Esegui il template con i dati delle lattine
    err = tmpl.Execute(w, pageData)
    if err != nil {
        pageData.ErrorMessage = "Errore nell'esecuzione del template: " + err.Error()
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
        if err != nil && err != http.ErrMissingFile {
            http.Error(w, "Errore nell'upload dell'immagine", http.StatusInternalServerError)
            return
        }

        // Salva l'immagine (se presente)
        var imagePath string
        if err == nil {
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

        // Aggiungi la lattina al database
        newCan := MonsterCan{Name: name, Flavor: flavor, Image: imagePath, Barcode: barcode}
        err = saveCans(newCan)
        if err != nil {
            http.Error(w, "Errore nel salvare i dati", http.StatusInternalServerError)
            return
        }

        // Redirect alla pagina principale dopo l'aggiunta
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // Carica il template per il form di aggiunta
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

func main() {
    // Gestisci le rotte
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/add", addHandler)
    // http.HandleFunc("/remove", removeHandler)

    // Servi i file statici
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // Avvia il server sulla porta 8080
    fmt.Println("Server in esecuzione su http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Errore nell'avvio del server:", err)
    }
}