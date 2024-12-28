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
    Can            MonsterCan
    SuccessMessage string
    ErrorMessage   string
}

// Connessione al database SQLite
func openDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./cans.db")
    if err != nil {
        return nil, err
    }

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

// Carica una lattina dal database
func loadCan(barcode string) (MonsterCan, error) {
    db, err := openDB()
    if err != nil {
        return MonsterCan{}, err
    }
    defer db.Close()

    var can MonsterCan
    err = db.QueryRow("SELECT name, flavor, image, barcode FROM cans WHERE barcode = ?", barcode).Scan(&can.Name, &can.Flavor, &can.Image, &can.Barcode)
    if err != nil {
        return MonsterCan{}, err
    }

    return can, nil
}

// Aggiungi una lattina al database
func saveCans(can MonsterCan) error {
    db, err := openDB()
    if err != nil {
        return err
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO cans (name, flavor, image, barcode) VALUES (?, ?, ?, ?)", can.Name, can.Flavor, can.Image, can.Barcode)
    return err
}

// Aggiorna una lattina nel database
func updateCan(can MonsterCan) error {
    db, err := openDB()
    if err != nil {
        return err
    }
    defer db.Close()

    _, err = db.Exec("UPDATE cans SET name = ?, flavor = ?, image = ?, barcode = ? WHERE barcode = ?", can.Name, can.Flavor, can.Image, can.Barcode, can.Barcode)
    return err
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

    if r.Method == http.MethodPost {
        if barcode := r.FormValue("barcode"); barcode != "" {
            if err := removeCan(barcode); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }
    }

    cans, err := loadCans()
    if err != nil {
        pageData.ErrorMessage = "Errore nel caricare i dati: " + err.Error()
    } else {
        pageData.Cans = cans
    }

    tmpl, err := template.ParseFiles("templates/index.html")
    if err != nil {
        pageData.ErrorMessage = "Errore nel caricare il template: " + err.Error()
    }

    if err := tmpl.Execute(w, pageData); err != nil {
        pageData.ErrorMessage = "Errore nell'esecuzione del template: " + err.Error()
    }
}

// Gestione della route per aggiungere una nuova lattina
func addHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        name := r.FormValue("name")
        flavor := r.FormValue("flavor")
        barcode := r.FormValue("barcode")
        image, _, err := r.FormFile("image")
        if err != nil && err != http.ErrMissingFile {
            http.Error(w, "Errore nell'upload dell'immagine", http.StatusInternalServerError)
            return
        }

        var imagePath string
        if err == nil {
            imagePath = "static/images/" + barcode + ".jpg"
            dst, err := os.Create(imagePath)
            if err != nil {
                http.Error(w, "Errore nel salvare l'immagine", http.StatusInternalServerError)
                return
            }
            defer dst.Close()
            if _, err := io.Copy(dst, image); err != nil {
                http.Error(w, "Errore nel copiare l'immagine", http.StatusInternalServerError)
                return
            }
        }

        newCan := MonsterCan{Name: name, Flavor: flavor, Image: imagePath, Barcode: barcode}
        if err := saveCans(newCan); err != nil {
            http.Error(w, "Errore nel salvare i dati", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    tmpl, err := template.ParseFiles("templates/add.html")
    if err != nil {
        http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := tmpl.Execute(w, nil); err != nil {
        http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
    }
}

// Gestione della route per modificare una lattina
func editHandler(w http.ResponseWriter, r *http.Request) {
    barcode := r.URL.Query().Get("barcode")
    if barcode == "" {
        http.Error(w, "Codice a barre mancante", http.StatusBadRequest)
        return
    }

    if r.Method == http.MethodPost {
        name := r.FormValue("name")
        flavor := r.FormValue("flavor")
        newBarcode := r.FormValue("barcode")
        image, _, err := r.FormFile("image")
        if err != nil && err != http.ErrMissingFile {
            http.Error(w, "Errore nell'upload dell'immagine", http.StatusInternalServerError)
            return
        }

        var imagePath string
        if err == nil {
            imagePath = "static/images/" + newBarcode + ".jpg"
            dst, err := os.Create(imagePath)
            if err != nil {
                http.Error(w, "Errore nel salvare l'immagine", http.StatusInternalServerError)
                return
            }
            defer dst.Close()
            if _, err := io.Copy(dst, image); err != nil {
                http.Error(w, "Errore nel copiare l'immagine", http.StatusInternalServerError)
                return
            }
        } else {
            can, err := loadCan(barcode)
            if err != nil {
                http.Error(w, "Errore nel caricare i dati della lattina", http.StatusInternalServerError)
                return
            }
            imagePath = can.Image
        }

        updatedCan := MonsterCan{Name: name, Flavor: flavor, Image: imagePath, Barcode: newBarcode}
        if err := updateCan(updatedCan); err != nil {
            http.Error(w, "Errore nel salvare i dati", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    can, err := loadCan(barcode)
    if err != nil {
        http.Error(w, "Errore nel caricare i dati della lattina", http.StatusInternalServerError)
        return
    }

    pageData := PageData{Can: can}
    tmpl, err := template.ParseFiles("templates/edit.html")
    if err != nil {
        http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := tmpl.Execute(w, pageData); err != nil {
        http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
    }
}

// Gestione della rotta per la pagina non trovata
func notFoundHandler(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNotFound)
    tmpl, err := template.ParseFiles("templates/404.html")
    if err != nil {
        http.Error(w, "Errore nel caricare il template: "+err.Error(), http.StatusInternalServerError)
        return
    }
    if err := tmpl.Execute(w, nil); err != nil {
        http.Error(w, "Errore nell'esecuzione del template: "+err.Error(), http.StatusInternalServerError)
    }
}

// Gestione delle rotte
func routeHandler(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/":
        homeHandler(w, r)
    case "/add":
        addHandler(w, r)
    case "/edit":
        editHandler(w, r)
    default:
        notFoundHandler(w)
    }
}

func main() {
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    http.HandleFunc("/", routeHandler)

    fmt.Println("Server in esecuzione su http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Errore nell'avvio del server:", err)
    }
}