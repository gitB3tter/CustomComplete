package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/docs"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	supportedMediaTypesStatus = mediatypes.Types{mediatypes.TextEventStream, mediatypes.HTML, mediatypes.JSON}
	//go:embed status.html
	templateStatusHTML string
	//go:embed partials/status.html
	templateStatusPartialsHTML string
)

type StatusSnapshot struct {
	StatusData
	ProgressData
}

type StatusData struct {
	Status  catalog.SourceStatus `json:"status"`
	Message string               `json:"message"`
}

type ProgressData struct {
	Progress int    `json:"progress"`
	Stage    string `json:"stage"`
}

func newStatusData(Status catalog.SourceStatus, Message string) StatusData {
	return StatusData{
		Status:  Status,
		Message: Message,
	}
}

func (h *handler) registerStatus(mux *http.ServeMux) {
	mux.Handle("GET /texts/{id}/status",
		transport.NoBody(h.status))
}

func (h *handler) initStatus(base *ui.BaseTemplates) (err error) {
	h.statusTemplate, err = base.BuildHTML(templateStatusHTML, templateStatusPartialsHTML)
	return
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())

	negotiated, ok := transport.NegotiateOr406(w, r, collectionLink(langTag), supportedMediaTypesStatus)
	if !ok {
		return
	}

	ctx := r.Context()

	T := i18n.Func(langTag)

	docID, doc, proceed := h.getDocOrRespond(w, r, r.PathValue("id"))
	if !proceed {
		return
	}

	itemLinkAbout := weblinking.New(weblinking.About, itemPath(docID)).WithTitle(string(doc.Name))

	//no more progress events updates
	if (doc.Status == catalog.StatusComplete || doc.Status == catalog.StatusFailed) && negotiated == mediatypes.TextEventStream {
		w.Header().Set("Link", itemLinkAbout.String())
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch negotiated {
	case mediatypes.HTML:
		output, err := ui.BufferHTML(h.statusTemplate[langTag], struct {
			Base ui.BaseDataHTML
			itemData
			Status StatusData
			ID     catalog.DocumentID
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationTextStatus, doc.Name),
				MetaDescription: T(i18n.TextStatusMetaDescription, doc.Name),
				Lang:            langTag,
				NoIndex:         true,
				Back:            itemLinkAbout,
				Links:           weblinking.Links{itemLinkAbout},
			},
			itemData: newItemData(doc),
			Status:   newStatusData(doc.Status, getStatusTranslation(doc.Status, langTag)),
		})
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
		transport.SetDontCache(w.Header())
		w.Header().Set("Link", itemLinkAbout.String())
		w.Header().Set("Content-Type", mediatypes.HTML.String())
		transport.Write(r.Context(), w, output)
		return
	case mediatypes.JSON:
		progress, err := h.service.GetDocumentProgressLatest(docID)
		var coded *errs.CodedError
		if err != nil && (!errors.As(err, &coded) || coded.Code() != errs.CodeNotFound) {
			transport.InternalError(w, r, err, itemLinkUp(docID, doc.Name))
			return
		}
		progressFound := err == nil
		snapshot := StatusSnapshot{
			StatusData: StatusData{
				Status:  doc.Status,
				Message: getStatusTranslation(doc.Status, langTag),
			},
		}
		if progressFound {
			snapshot.Progress = progress.Percentage
			snapshot.Stage = T(progress.Message, progress.ArgsMessage...)
		}
		body, err := json.Marshal(snapshot)
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
		transport.SetDontCache(w.Header())
		w.Header().Set("Link", weblinking.Links{
			itemLinkAbout,
			weblinking.New(weblinking.DescribedBy, (&url.URL{Path: "/"}).JoinPath(string(docs.SchemaTextsItemStatus))),
		}.String())
		w.Header().Set("Content-Type", mediatypes.JSON.String())
		transport.Write(ctx, w, body)
		return

	}

	updates, latest, hasLatest, unsub, err := h.service.SubscribeDocumentProgress(docID)
	if err != nil {
		transport.InternalError(w, r, err, itemLinkUp(docID, doc.Name))
		return
	}
	defer unsub()
	rc := http.NewResponseController(w)

	if err := rc.SetWriteDeadline(time.Now().Add(3 * time.Minute)); err != nil {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		transport.InternalError(w, r, err, itemLinkAbout)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	transport.SetNoCache(w.Header())
	w.Header().Set("Connection", "keep-alive")

	if err := rc.Flush(); err != nil {
		if errors.Is(err, http.ErrNotSupported) {
			return
		}
		respondInternalErrorSubresource(w, r, fmt.Errorf("error initiating ResponseController for SSE: %w", err), docID, doc.Name)
		return
	}

	buf := make([]byte, 0, 256)
	var lastPercentage int = -1
	var lastState i18n.Key

	anounceProcessingDelayed := doc.Status == catalog.StatusPending
	if !anounceProcessingDelayed {
		buf = append(buf, "data: "...)
		buf = append(buf, []byte(getStatusTranslation(catalog.StatusProcessing, langTag))...)
		buf = append(buf, "\n\n"...)
		buf = append(buf, "event: status\ndata: "...)
		buf = append(buf, []byte(catalog.StatusProcessing)...)
		buf = append(buf, "\n\n"...)
		if _, err := w.Write(buf); err != nil {
			return
		}
		if err := rc.Flush(); err != nil {
			return
		}
	}

	if hasLatest {
		fmt.Fprintf(w, "event: progress\ndata: %d\n\n", latest.Percentage)
		fmt.Fprintf(w, "event: stage\ndata: %s\n\n", T(latest.Message, latest.ArgsMessage...))
		lastPercentage = latest.Percentage
		lastState = latest.Message
		if _, err := w.Write(buf); err != nil {
			return
		}
		if err := rc.Flush(); err != nil {
			return
		}
	}

	clientGone := r.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case latest, ok := <-updates:
			if !ok {
				h.respondWithFinalStatus(ctx, w, rc, docID)
				return
			}

			buf = buf[:0]

			if anounceProcessingDelayed {
				buf = append(buf, "data: "...)
				buf = append(buf, []byte(getStatusTranslation(catalog.StatusProcessing, langTag))...)
				buf = append(buf, "\n\n"...)
				buf = append(buf, "event: status\ndata: "...)
				buf = append(buf, []byte(catalog.StatusProcessing)...)
				buf = append(buf, "\n\n"...)
				anounceProcessingDelayed = false
			}

			if latest.Percentage == 100 {
				if _, err := w.Write(buf); err != nil {
					return
				}
				if err := rc.Flush(); err != nil {
					return
				}
				h.respondWithFinalStatus(ctx, w, rc, docID)
				return
			}

			if latest.Percentage != lastPercentage {
				buf = append(buf, "event: progress\ndata: "...)
				buf = strconv.AppendInt(buf, int64(latest.Percentage), 10)
				buf = append(buf, "\n\n"...)
			}

			if latest.Message != lastState {
				buf = append(buf, "event: stage\ndata: "...)
				buf = append(buf, []byte(T(latest.Message, latest.ArgsMessage...))...)
				buf = append(buf, "\n\n"...)
			}

			if _, err := w.Write(buf); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}

			lastPercentage = latest.Percentage
			lastState = latest.Message
		}
	}
}

func (h *handler) respondWithFinalStatus(ctx context.Context, w http.ResponseWriter, rc *http.ResponseController, docID catalog.DocumentID) {
	buf := make([]byte, 0, 256)
	doc, err := h.service.GetDocument(docID)
	if err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("failed getting document for final status: %v", err))

		buf = append(buf, "event: status\ndata: "...)
		buf = append(buf, []byte("unknown")...)
		buf = append(buf, "\n\n"...)
		buf = append(buf, "event: done\ndata: "...)
		buf = append(buf, []byte(i18n.TranslateFuncFromContext(ctx)(i18n.TextStatusUnknown))...)
		buf = append(buf, "\n\n"...)
		if _, err := w.Write(buf); err != nil {
			return
		}
		if err := rc.Flush(); err != nil {
			return
		}
		return
	}

	var percentage int64 = 0
	if doc.Status == catalog.StatusComplete {
		percentage = 100
	}

	buf = append(buf, "event: progress\ndata: "...)
	buf = strconv.AppendInt(buf, percentage, 10)
	buf = append(buf, "\n\n"...)

	buf = append(buf, "event: status\ndata: "...)
	buf = append(buf, []byte(doc.Status)...)
	buf = append(buf, "\n\n"...)
	buf = append(buf, "event: done\ndata: "...)
	buf = append(buf, []byte(getStatusTranslation(doc.Status, i18n.LangTagFromContext(ctx)))...)
	buf = append(buf, "\n\n"...)

	if _, err := w.Write(buf); err != nil {
		return
	}
	if err := rc.Flush(); err != nil {
		return
	}

}
