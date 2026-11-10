package repositorio

import (
	"fmt"
	"sync"

	"audios.local/servidor-audios/modelo"
)

type RepositorioAudios struct {
	mu          sync.RWMutex
	audios      map[string]modelo.Audio
	siguienteID int
}

func NuevoRepositorio() *RepositorioAudios {
	r := &RepositorioAudios{audios: make(map[string]modelo.Audio)}
	r.cargarDatosIniciales()
	return r
}

func (r *RepositorioAudios) Guardar(a modelo.Audio) modelo.Audio {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.siguienteID++
	a.ID = fmt.Sprintf("a%d", r.siguienteID)
	r.audios[a.ID] = a
	return a
}

func (r *RepositorioAudios) ListarPorTipo(tipo string) []modelo.Audio {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := []modelo.Audio{}
	for _, a := range r.audios {
		if a.Tipo == tipo {
			resultado = append(resultado, a)
		}
	}
	return resultado
}

func (r *RepositorioAudios) BuscarPorID(id string) (modelo.Audio, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, existe := r.audios[id]
	return a, existe
}

func (r *RepositorioAudios) cargarDatosIniciales() {
	r.Guardar(modelo.Audio{
		Titulo: "Canción de prueba 1", Tipo: "musica", Archivo: "cancion1", Formato: "mp3",
		Metadatos: map[string]string{"artista": "Artista Demo", "album": "Álbum Demo", "genero": "Pop"},
	})
	r.Guardar(modelo.Audio{
		Titulo: "Canción de prueba 2", Tipo: "musica", Archivo: "cancion2", Formato: "mp3",
		Metadatos: map[string]string{"artista": "Banda Demo", "album": "Sencillo", "genero": "Rock"},
	})
}
