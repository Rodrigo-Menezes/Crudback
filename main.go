package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"google.golang.org/api/option"
)

var (
	firebaseDB *db.Client
)

type Item struct {
	ID       string `json:"id"`
	Nome     string `json:"nome"`
	Telefone int    `json:"telefone"`
	Endereco string `json:"endereco"`
}

func main() {
	// Configura as opções para o Firebase com o arquivo de credenciais
	opt := option.WithCredentialsFile("config/auth.json")

	// Configura as opções do Firebase com a URL do Realtime Database
	config := &firebase.Config{
		DatabaseURL: "https://crudemnextgo-default-rtdb.firebaseio.com/",
	}

	// Inicializa o aplicativo Firebase com as opções
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	// Inicializa o cliente Firebase Realtime Database
	firebaseDB, err = app.Database(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firebase DB client: %v\n", err)
	}

	// Crie um roteador usando a biblioteca Gorilla Mux
	r := mux.NewRouter()

	// Configuração do CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"crud-front-delta.vercel.app"}, // Altere isso para o domínio do seu aplicativo React em produção
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	// Aplica o middleware de CORS ao roteador
	r.Use(corsHandler.Handler)

	// Defina suas rotas aqui usando o roteador "r"
	r.HandleFunc("/create", createItem).Methods("POST")
	r.HandleFunc("/read", readItems).Methods("GET")
	r.HandleFunc("/update", updateItem).Methods("PUT")
	r.HandleFunc("/delete", deleteItem).Methods("DELETE")

	// Obtém a porta a partir das variáveis de ambiente (ou usa 8080 como padrão)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Inicia o servidor HTTP na porta especificada
	fmt.Printf("Server is listening on port %s...\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

// função responsalvel por criar items no firebase
func createItem(w http.ResponseWriter, r *http.Request) {
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Obtém uma referência à coleção "items" no Firebase Realtime Database
	ref := firebaseDB.NewRef("items")

	// Cria um novo nó na coleção "items" e obtém a referência para o novo nó e um erro, se houver
	childRef, err := ref.Push(context.Background(), newItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Define o ID do novo item com a chave gerada pelo Firebase
	newItem.ID = childRef.Key

	// Define o código de status HTTP 201 Created e envia o novo item como resposta
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

// função responsalvel por ler os item no firebase
func readItems(w http.ResponseWriter, r *http.Request) {
	// Obtém uma referência à coleção "items" no Firebase Realtime Database
	ref := firebaseDB.NewRef("items")

	// Declara uma variável para armazenar os dados do Firebase
	var data map[string]Item

	// Obtém os dados do Firebase e os armazena na variável "data"
	if err := ref.Get(context.Background(), &data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Converte o mapa de dados em uma fatia de objetos "Item"
	var items []Item
	for _, v := range data {
		items = append(items, v)
	}

	// Codifica os itens em JSON e os envia como resposta
	json.NewEncoder(w).Encode(items)
}

// funcção responsavel por atualizar o item no firebase
func updateItem(w http.ResponseWriter, r *http.Request) {
	// Decodifique os dados da solicitação JSON em um objeto Item com os campos atualizados
	var updatedItem Item
	err := json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Obtenha o ID do item que você deseja atualizar a partir dos parâmetros da URL
	itemID := r.URL.Query().Get("itemID")

	// Verifique se o ID do item é válido
	if itemID == "" {
		http.Error(w, "Item ID is missing", http.StatusBadRequest)
		return
	}

	// Obtenha uma referência ao nó específico que deseja atualizar no Firebase
	ref := firebaseDB.NewRef("items").Child(itemID)

	// Atualize os campos relevantes do item no Firebase
	if err := ref.Update(context.Background(), map[string]interface{}{
		"nome":     updatedItem.Nome,
		"telefone": updatedItem.Telefone,
		"endereco": updatedItem.Endereco,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Responda com uma confirmação de sucesso
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Item with ID %s updated successfully", itemID)
}

// Função responsavel por deletar items do firebase
func deleteItem(w http.ResponseWriter, r *http.Request) {
	// Obtenha o ID do item que você deseja excluir dos parâmetros da URL
	itemID := r.URL.Query().Get("itemID")

	// Verifique se o ID do item é válido
	if itemID == "" {
		http.Error(w, "Item ID is missing", http.StatusBadRequest)
		return
	}

	// Obtém uma referência ao nó específico que deseja excluir no Firebase
	ref := firebaseDB.NewRef("items").Child(itemID)

	// Remove o nó do Firebase
	if err := ref.Delete(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Responda com uma confirmação de sucesso
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Item with ID %s deleted successfully", itemID)
}
