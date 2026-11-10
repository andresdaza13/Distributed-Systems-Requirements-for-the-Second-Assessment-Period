package fachada

import (
	"errors"

	"audios.local/servidor-audios/modelo"
	"audios.local/servidor-audios/repositorio"
)

var (
	ErrTipoInvalido = errors.New("tipo de audio no válido")
	ErrNoEncontrado = errors.New("audio no encontrado")
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
