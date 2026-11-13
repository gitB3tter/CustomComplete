package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/schema"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/text/language"
)

var (
	supportedMediaTypesStatus = mediatypes.Types{mediatypes.TextEventStream, mediatypes.HTML, mediatypes.JSON}
	//go:embed status.html
	templateStatusHTML string
	//go:embed partials/status.html
	templateStatusPartialsHTML string
	StatusSchemaPath           weblinking.Path = ItemSchemaPath.Join("status")
)

type StatusSnapshot struct {
	StatusData
	ProgressData
}

type StatusData struct {
	Status  catalog.SourceStatus `json:"status"`
	Message string               `json:"message"`
}

func StatusSchema(langTag language.Tag) schema.Root {
	T := i18n.Func(langTag)

	required := map[string]schema.Schema{
		"status": schema.Schema{Type: schema.String}.
			WithEnum(
				catalog.StatusPending,
				catalog.StatusProcessing,
				catalog.StatusComplete,
				catalog.StatusFailed,
			),
		"message": schema.Schema{Type: schema.String}.
			WithExamples(
				T(i18n.TextStatusPending),
				T(i18n.TextStatusProcessing),
				T(i18n.TextStatusComplete),
				T(i18n.TextStatusFailed),
			),
	}

	optional := map[string]schema.Schema{
		"progress": schema.New(schema.Integer).
			WithMinimum(0).
			WithMaximum(100),
		"stage": schema.New(schema.String).
			WithExamples(
				T(i18n.ProcessingReadingText),
				T(i18n.ProcessingCounting),
				T(i18n.ProcessingPreparingSuggestions),
				T(i18n.ProcessingSavingSuggestions),
			),
	}

	return schema.NewRoot(StatusSchemaPath, required).
		WithTitle(T(i18n.NavigationTextStatus, T(i18n.DocsSpecsExampleTextName))).
		WithOptional(optional)
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

func TextsStatusPath(docID catalog.DocumentID) weblinking.Path {
	return TextsItemPath(docID).Join("status")
}

func (h *handler) registerStatus(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/status",
		transport.NoBody(h.status))
}

func (h *handler) initStatus(base *ui.BaseTemplates) (err error) {
	h.statusTemplate, err = base.BuildHTML(templateStatusHTML, templateStatusPartialsHTML)
	return
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())

	responseType, ok := transport.NegotiateOr406(w, r, collectionLink(langTag), supportedMediaTypesStatus)
	if !ok {
		return
	}
	ctx := r.Context()

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	T := i18n.Func(langTag)

	//no more progress events updates
	if (doc.Status == catalog.StatusComplete || doc.Status == catalog.StatusFailed) && responseType == mediatypes.TextEventStream {
		w.Header().Set("Link", itemLinkAbout(docID, doc.Name).String())
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch responseType {
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
				Back:            itemLinkAbout(docID, doc.Name),
				Links:           weblinking.Links{itemLinkAbout(docID, doc.Name)},
			},
			itemData: newItemData(doc),
			Status:   newStatusData(doc.Status, getStatusTranslation(doc.Status, langTag)),
		})
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
		transport.SetDontCache(w.Header())
		w.Header().Set("Link", itemLinkAbout(docID, doc.Name).String())
		transport.Write(r.Context(), w, output, responseType, http.StatusOK)
		return
	case mediatypes.JSON:
		snapshot := StatusSnapshot{
			StatusData: StatusData{
				Status:  doc.Status,
				Message: getStatusTranslation(doc.Status, langTag),
			},
		}
		if progress, err := h.service.GetDocumentProgressLatest(docID); err == nil {
			snapshot.Progress = progress.Percentage
			snapshot.Stage = T(progress.Message, progress.ArgsMessage...)
		} else if !errs.HasCode(err, errs.CodeNotFound) {
			transport.InternalError(w, r, err, itemLinkUp(docID, doc.Name))
			return
		}
		body, err := json.Marshal(snapshot)
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
		transport.SetDontCache(w.Header())
		w.Header().Set("Link", weblinking.Links{
			itemLinkAbout(docID, doc.Name),
			weblinking.New(weblinking.DescribedBy, StatusSchemaPath),
		}.String())
		transport.Write(ctx, w, body, responseType, http.StatusOK)
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
		transport.InternalError(w, r, err, itemLinkAbout(docID, doc.Name))
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
		buf = append(buf, getStatusTranslation(catalog.StatusProcessing, langTag)...)
		buf = append(buf, "\n\n"...)
		buf = append(buf, "event: status\ndata: "...)
		buf = append(buf, catalog.StatusProcessing...)
		buf = append(buf, "\n\n"...)
		if _, err := w.Write(buf); err != nil {
			return
		}
		if err := rc.Flush(); err != nil {
			return
		}
	}

	if hasLatest {
		buf = buf[:0]
		buf = append(buf, "event: progress\ndata: "...)
		buf = strconv.AppendInt(buf, int64(latest.Percentage), 10)
		buf = append(buf, "\n\nevent: stage\ndata: "...)
		buf = append(buf, T(latest.Message, latest.ArgsMessage...)...)
		buf = append(buf, "\n\n"...)
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
				buf = append(buf, getStatusTranslation(catalog.StatusProcessing, langTag)...)
				buf = append(buf, "\n\n"...)
				buf = append(buf, "event: status\ndata: "...)
				buf = append(buf, catalog.StatusProcessing...)
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
				buf = append(buf, T(latest.Message, latest.ArgsMessage...)...)
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
		buf = append(buf, "unknown"...)
		buf = append(buf, "\n\n"...)
		buf = append(buf, "event: done\ndata: "...)
		buf = append(buf, i18n.TranslateFuncFromContext(ctx)(i18n.TextStatusUnknown)...)
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
	buf = append(buf, doc.Status...)
	buf = append(buf, "\n\n"...)
	buf = append(buf, "event: done\ndata: "...)
	buf = append(buf, getStatusTranslation(doc.Status, i18n.LangTagFromContext(ctx))...)
	buf = append(buf, "\n\n"...)

	if _, err := w.Write(buf); err != nil {
		return
	}
	if err := rc.Flush(); err != nil {
		return
	}

}
