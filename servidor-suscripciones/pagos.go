package main

import (
	"log"
	"sync"
)

type Plan struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Precio int    `json:"precio"`
}

var Planes = []Plan{
	{ID: "basico", Nombre: "Plan Básico", Precio: 10000},
	{ID: "premium", Nombre: "Plan Premium", Precio: 20000},
	{ID: "familiar", Nombre: "Plan Familiar", Precio: 35000},
}

type Pago struct {
	Usuario string `json:"usuario"`
	Plan    string `json:"plan"`
	Monto   int    `json:"monto"`
}

type Resultado struct {
	Estado     string `json:"estado"`
	Plan       string `json:"plan"`
	Monto      int    `json:"monto"`
	Almacenado bool   `json:"almacenado"`
	Repetido   bool   `json:"repetido"`
	Token      string `json:"token"`
	Mensaje    string `json:"mensaje"`
}

// FachadaPagos implementa la idempotencia: token -> resultado.
type FachadaPagos struct {
	mu         sync.Mutex
	procesados map[string]Resultado
	historial  map[string][]Resultado
}

func NuevaFachadaPagos() *FachadaPagos {
	return &FachadaPagos{
		procesados: make(map[string]Resultado),
		historial:  make(map[string][]Resultado),
	}
}

// Pagar es atómica gracias al mutex: "verificar y guardar" sin carreras.
// Dos reintentos simultáneos con el mismo token jamás cobran dos veces.
func (f *FachadaPagos) Pagar(token string, p Pago) (Resultado, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r, existe := f.procesados[token]; existe {
		log.Printf("ECO servidor: token %s YA procesado -> se devuelve el resultado guardado, NO se cobra de nuevo", token)
		r.Repetido = true
		r.Mensaje = "token ya utilizado: no se realizó un nuevo cobro"
		return r, true
	}

	log.Printf("ECO servidor: token %s recibido por primera vez, procesando pago...", token)
	r := Resultado{
		Estado: "exitoso", Plan: p.Plan, Monto: p.Monto,
		Almacenado: true, Token: token,
		Mensaje: "pago procesado y almacenado",
	}
	f.procesados[token] = r
	f.historial[p.Usuario] = append(f.historial[p.Usuario], r)
	log.Printf("ECO servidor: pago de '%s' por el plan '%s' almacenado (token %s)", p.Usuario, p.Plan, token)
	return r, false
}

func (f *FachadaPagos) HistorialDe(usuario string) []Resultado {
	f.mu.Lock()
	defer f.mu.Unlock()
	resultado := f.historial[usuario]
	if resultado == nil {
		resultado = []Resultado{}
	}
	return resultado
}
