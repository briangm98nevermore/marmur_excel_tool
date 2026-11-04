package model

import "time"

type Entry struct {
	Date        time.Time
	Type        string   // "Entrada" o "Gasto"
	Category    string   // nomenclador
	InvoiceNum  string
	Description string
	Amount      float64
	Currency    string   // "USD", "CUP", "EUR"
}

// Nomencladores iniciales
var DefaultIncomeCategories = []string{"Saldo inicial", "Empresa", "Cambios", "Créditos", "Otros"}
var DefaultExpenseCategories = []string{"Casa", "Almacén", "Irma Administración", "Pagos", "Otros"}