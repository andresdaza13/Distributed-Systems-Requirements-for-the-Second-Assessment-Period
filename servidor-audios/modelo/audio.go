package modelo

// Audio representa cualquier audio del catálogo, sin importar su tipo.
type Audio struct {
	ID        string            `json:"id"`
	Titulo    string            `json:"titulo"`
	Tipo      string            `json:"tipo"`
	Archivo   string            `json:"archivo"` // nombre SIN extensión, ej: "cancion1"
	Formato   string            `json:"formato"` // ej: "mp3"
	Metadatos map[string]string `json:"metadatos"`
}

// Tipos válidos, en el orden en que se mostrarán al usuario.
var Tipos = []string{"musica", "podcast", "audiolibro", "ruido_blanco"}

// CamposPorTipo define qué metadatos obligatorios tiene cada tipo (esquema).
var CamposPorTipo = map[string][]string{
	"musica":       {"artista", "album", "genero"},
	"podcast":      {"nombre_podcast", "titulo_episodio", "anfitrion"},
	"audiolibro":   {"titulo_libro", "autor", "narrador"},
	"ruido_blanco": {"tipo_sonido", "fuente", "uso_sugerido"},
}
