package excel

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxUpload = 10 << 20 // 10mb

type Handler struct {
	svc    *Service
	errors errores.RespuestaErrores
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	file, hdr, err := r.FormFile("file")

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, errors.New("archivo requerido o demasiado grande"))
		return
	}

	defer file.Close()

	magic := make([]byte, 2)
	if _, err := io.ReadFull(file, magic); err != nil || string(magic) != "PK" {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	ctx := r.Context()

	userID := utilidades.ObtenerUsuarioDelContexto(r).ID
	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	rec, err := h.svc.Save(ctx, userID, empresaID, filepath.Base(hdr.Filename), file, hdr.Size)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.RespuestaJSON(w, http.StatusOK, rec); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	files, err := h.svc.List(ctx, empresaID, limit, offset)

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	if err := utilidades.EscribirJSON(w, http.StatusOK, files); err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r, "id"))
	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	f, rc, err := h.svc.Open(r.Context(), empresaID, ID)

	if errors.Is(err, ErrNotFound) {
		h.errors.ErrorNoEncontrado(w, r, err)
		return
	}

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	defer rc.Close()

	w.Header().Set("Content-Type", xlsxType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": f.Nombre}))
	w.Header().Set("Content-Length", strconv.FormatInt(f.SizeBytes, 10))
	_, _ = io.Copy(w, rc)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(chi.URLParam(r, "id"))
	empresaID := utilidades.ObtenerEmpresaDelContexto(r).ID

	if err != nil {
		h.errors.ErrorSolicitudInvalida(w, r, err)
		return
	}

	err = h.svc.Delete(r.Context(), empresaID, ID)

	if errors.Is(err, ErrNotFound) {
		h.errors.ErrorNoEncontrado(w, r, err)
		return
	}

	if err != nil {
		h.errors.ErrorInterno(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
