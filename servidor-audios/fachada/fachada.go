package fachada

import (
	"errors"
	"fmt"
	"strings"

	"audios.local/servidor-audios/modelo"
	"audios.local/servidor-audios/repositorio"
)

var (
	ErrTipoInvalido         = errors.New("tipo de audio no válido")
	ErrNoEncontrado         = errors.New("audio no encontrado")
	ErrMetadatosIncompletos = errors.New("metadatos incompletos")
)

type FachadaAudios struct {
	repo *repositorio.RepositorioAudios
}

func NuevaFachada(repo *repositorio.RepositorioAudios) *FachadaAudios {
	return &FachadaAudios{repo: repo}
}

func (f *FachadaAudios) ListarTipos() []string {
	return modelo.Tipos
}

func (f *FachadaAudios) CamposDeTipo(tipo string) ([]string, error) {
	campos, valido := modelo.CamposPorTipo[tipo]
	if !valido {
		return nil, ErrTipoInvalido
	}
	return campos, nil
}

func (f *FachadaAudios) ListarPorTipo(tipo string) ([]modelo.Audio, error) {
	if _, valido := modelo.CamposPorTipo[tipo]; !valido {
		return nil, ErrTipoInvalido
	}
	return f.repo.ListarPorTipo(tipo), nil
}

func (f *FachadaAudios) ObtenerDetalle(id string) (modelo.Audio, error) {
	audio, existe := f.repo.BuscarPorID(id)
	if !existe {
		return modelo.Audio{}, ErrNoEncontrado
	}
	return audio, nil
}

// Registrar valida los metadatos contra el esquema del tipo (opción f).
// La validación del servidor es obligatoria: el cliente no es confiable.
func (f *FachadaAudios) Registrar(a modelo.Audio) (modelo.Audio, error) {
	campos, valido := modelo.CamposPorTipo[a.Tipo]
	if !valido {
		return modelo.Audio{}, ErrTipoInvalido
	}
	if strings.TrimSpace(a.Titulo) == "" {
		return modelo.Audio{}, fmt.Errorf("%w: falta el título", ErrMetadatosIncompletos)
	}
	for _, campo := range campos {
		if strings.TrimSpace(a.Metadatos[campo]) == "" {
			return modelo.Audio{}, fmt.Errorf("%w: falta '%s' para el tipo %s", ErrMetadatosIncompletos, campo, a.Tipo)
		}
	}
	return f.repo.Guardar(a), nil
}
