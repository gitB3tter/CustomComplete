package transport

import (
	"crypto/sha256"
	"customcomplete/internal/httpsyntax"
	"customcomplete/internal/i18n"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

type ETagWeak string

type ETagStrong string

func GetEtagWeak(output []byte) ETagWeak {
	hash := sha256.Sum256(output)
	return ETagWeak(fmt.Sprintf(`W/"%x"`, hash[:16]))
}

// generates a 66 bytes long Strong ETag (sha-256: 32 bytes =%x=> 64 bytes + 2x" => 66 bytes )
func GetEtagStrong(v any) (ETagStrong, error) {
	hash := sha256.New()
	enc := gob.NewEncoder(hash)
	if err := enc.Encode(v); err != nil {
		return "", fmt.Errorf("failed to encode v(%v) for ETag: %w", v, err)
	}
	sum := hash.Sum(nil)

	return ETagStrong(fmt.Sprintf(`"%x"`, sum)), nil
}

func GetEtagStrongBytes(b []byte) ETagStrong {
	return ETagStrong(fmt.Sprintf(`"%x"`, sha256.Sum256(b)))
}

func GetEtagStrongLiteral(str string) ETagStrong {
	return ETagStrong(`"` + str + `"`)
}

func normalizeETag(etag string) string {
	return strings.Trim(strings.TrimPrefix(strings.TrimSpace(etag), "W/"), `"`)
}

func ETagMatchWeak(clientEtags []string, etagWeak ETagWeak) bool {
	serverEtagNorm := normalizeETag(string(etagWeak))

	for _, header := range clientEtags {
		for _, val := range httpsyntax.SplitList(header, ',') {
			if strings.TrimSpace(val) == "*" {
				return true
			}

			if normalizeETag(val) == serverEtagNorm {
				return true
			}
		}
	}
	return false
}

func EtagMatchStrong(clientEtags []string, etagStrong ETagStrong) bool {

	serverEtagNorm := normalizeETag(string(etagStrong))

	for _, header := range clientEtags {
		for _, val := range httpsyntax.SplitList(header, ',') {
			if strings.TrimSpace(val) == "*" {
				return true
			}

			if strings.HasPrefix(val, "W/") {
				continue
			}

			if normalizeETag(val) == serverEtagNorm {
				return true
			}
		}
	}
	return false
}

func Precondition(etagStrong ETagStrong, ETagClient string, langTag language.Tag) *ErrorPrecondition {
	if ETagClient == "" {
		return &ErrorPrecondition{
			status:  http.StatusPreconditionRequired,
			langTag: langTag,
		}
	}
	if !EtagMatchStrong([]string{ETagClient}, etagStrong) {
		return &ErrorPrecondition{
			status:  http.StatusPreconditionFailed,
			langTag: langTag,
		}
	}
	return nil
}

func ErrPrecondition(status int, langTag language.Tag) ErrorPrecondition {
	return ErrorPrecondition{
		status:  status,
		langTag: langTag,
	}
}

func RespondNoContent(w http.ResponseWriter) {
	w.Header().Del("Vary")
	w.Header().Del("Content-Language")
	w.WriteHeader(http.StatusNoContent)
}

type ErrorPrecondition struct {
	status  int
	langTag language.Tag
}

func (pre *ErrorPrecondition) Status() int {
	return pre.status
}

func (pre *ErrorPrecondition) Error() string {
	T := i18n.Func(pre.langTag)
	if pre.status == http.StatusPreconditionRequired {
		return T(i18n.HTTPErrorPreconditionRequiredDescription)
	}
	return T(i18n.HTTPErrorPreconditionFailedDescription)
}

func (pre *ErrorPrecondition) String() string {
	return pre.Error()
}

func (pre *ErrorPrecondition) Required() bool {
	return pre.status == http.StatusPreconditionRequired
}

func (pre *ErrorPrecondition) Failed() bool {
	return pre.status == http.StatusPreconditionFailed
}

func (pre *ErrorPrecondition) CustomMessage(failed string, required string) string {
	if pre.status == http.StatusPreconditionRequired {
		return required
	}
	return failed
}
