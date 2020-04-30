package utils

import (
	"github.com/yacchi/lambda-http-adaptor/types"
	"mime"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var binaryContentEncoding = map[string]struct{}{
	"gzip":    {},
	"x-gzip":  {},
	"deflate": {},
	"br":      {},
}

func RegisterBinaryContentEncoding(t string) {
	binaryContentEncoding[t] = struct{}{}
}

var textMIMERegexp = []*regexp.Regexp{
	regexp.MustCompile("^text/"),
}

var textMIMEType = map[string]struct{}{
	"image/svg+xml":          {},
	"application/json":       {},
	"application/javascript": {},
	"application/xml":        {},
}

func RegisterTextMIMEType(t string) {
	textMIMEType[t] = struct{}{}
}

func RegisterTextMIMERegexp(r *regexp.Regexp) {
	textMIMERegexp = append(textMIMERegexp, r)
}

func IsBinaryContent(h http.Header) bool {
	for _, t := range h.Values(types.HTTPHeaderContentEncoding) {
		if _, ok := binaryContentEncoding[t]; ok {
			return true
		}
	}

	if IsTextContent(h.Get(types.HTTPHeaderContentType)) {
		return false
	}
	return true
}

func IsTextContent(contentType string) bool {
	m, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	_, ok := textMIMEType[m]

	if ok {
		return true
	}

	for _, re := range textMIMERegexp {
		if re.MatchString(m) {
			textMIMEType[m] = struct{}{}
			return true
		}
	}

	return false
}

func SemicolonSeparatedHeaderMap(h http.Header) map[string]string {
	ret := map[string]string{}
	for k, v := range h {
		ret[k] = strings.Join(v, ";")
	}
	return ret
}

func JoinMultiValueQueryParameters(param map[string][]string) string {
	if param == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := param[k]
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func JoinQueryParameters(param map[string]string) string {
	if param == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(param))
	for k := range param {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := param[k]
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
	}
	return buf.String()
}
