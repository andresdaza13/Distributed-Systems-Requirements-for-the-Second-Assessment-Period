package main

import "sync"

type Evento struct {
	Usuario string `json:"usuario"`
	AudioID string `json:"audioId"`
}

type ItemHistorial struct {
	AudioID string `json:"audioId"`
	Veces   int    `json:"veces"`
}

// Contador es el "repositorio" de este servicio: usuario -> audio -> veces.
type Contador struct {
	mu    sync.Mutex
	datos map[string]map[string]int
}

func NuevoContador() *Contador {
	return &Contador{datos: make(map[string]map[string]int)}
}

func (c *Contador) Sumar(usuario, audioID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.datos[usuario] == nil {
		c.datos[usuario] = make(map[string]int)
	}
	c.datos[usuario][audioID]++
}

func (c *Contador) Listar(usuario string) []ItemHistorial {
	c.mu.Lock()
	defer c.mu.Unlock()
	resultado := []ItemHistorial{}
	for audioID, veces := range c.datos[usuario] {
		resultado = append(resultado, ItemHistorial{AudioID: audioID, Veces: veces})
	}
	return resultado
}
