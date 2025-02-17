package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AstaFrode/boxo/ipns"
	"github.com/AstaFrode/boxo/routing/http/internal/drjson"
	"github.com/AstaFrode/boxo/routing/http/types"
	"github.com/AstaFrode/boxo/routing/http/types/iter"
	jsontypes "github.com/AstaFrode/boxo/routing/http/types/json"
	"github.com/AstaFrode/go-libp2p/core/peer"
	"github.com/cespare/xxhash/v2"
	"github.com/gorilla/mux"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"

	logging "github.com/ipfs/go-log/v2"
)

const (
	mediaTypeJSON       = "application/json"
	mediaTypeNDJSON     = "application/x-ndjson"
	mediaTypeWildcard   = "*/*"
	mediaTypeIPNSRecord = "application/vnd.ipfs.ipns-record"

	DefaultRecordsLimit          = 20
	DefaultStreamingRecordsLimit = 0
)

var logger = logging.Logger("service/server/delegatedrouting")

const (
	ProvidePath       = "/routing/v1/providers/"
	FindProvidersPath = "/routing/v1/providers/{cid}"
	IPNSPath          = "/routing/v1/ipns/{cid}"
)

type FindProvidersAsyncResponse struct {
	ProviderResponse types.ProviderResponse
	Error            error
}

type ContentRouter interface {
	// FindProviders searches for peers who are able to provide a given key. Limit
	// indicates the maximum amount of results to return. 0 means unbounded.
	FindProviders(ctx context.Context, key cid.Cid, limit int) (iter.ResultIter[types.ProviderResponse], error)
	ProvideBitswap(ctx context.Context, req *BitswapWriteProvideRequest) (time.Duration, error)
	Provide(ctx context.Context, req *WriteProvideRequest) (types.ProviderResponse, error)

	// FindIPNSRecord searches for an [ipns.Record] for the given [ipns.Name].
	FindIPNSRecord(ctx context.Context, name ipns.Name) (*ipns.Record, error)

	// ProvideIPNSRecord stores the provided [ipns.Record] for the given [ipns.Name]. It is
	// guaranteed that the record matches the provided name.
	ProvideIPNSRecord(ctx context.Context, name ipns.Name, record *ipns.Record) error
}

type BitswapWriteProvideRequest struct {
	Keys        []cid.Cid
	Timestamp   time.Time
	AdvisoryTTL time.Duration
	ID          peer.ID
	Addrs       []multiaddr.Multiaddr
}

type WriteProvideRequest struct {
	Protocol string
	Schema   string
	Bytes    []byte
}

type Option func(s *server)

// WithStreamingResultsDisabled disables ndjson responses, so that the server only supports JSON responses.
func WithStreamingResultsDisabled() Option {
	return func(s *server) {
		s.disableNDJSON = true
	}
}

// WithRecordsLimit sets a limit that will be passed to ContentRouter.FindProviders
// for non-streaming requests (application/json). Default is DefaultRecordsLimit.
func WithRecordsLimit(limit int) Option {
	return func(s *server) {
		s.recordsLimit = limit
	}
}

// WithStreamingRecordsLimit sets a limit that will be passed to ContentRouter.FindProviders
// for streaming requests (application/x-ndjson). Default is DefaultStreamingRecordsLimit.
func WithStreamingRecordsLimit(limit int) Option {
	return func(s *server) {
		s.streamingRecordsLimit = limit
	}
}

func Handler(svc ContentRouter, opts ...Option) http.Handler {
	server := &server{
		svc:                   svc,
		recordsLimit:          DefaultRecordsLimit,
		streamingRecordsLimit: DefaultStreamingRecordsLimit,
	}

	for _, opt := range opts {
		opt(server)
	}

	r := mux.NewRouter()
	r.HandleFunc(ProvidePath, server.provide).Methods(http.MethodPut)
	r.HandleFunc(FindProvidersPath, server.findProviders).Methods(http.MethodGet)

	r.HandleFunc(IPNSPath, server.getIPNSRecord).Methods(http.MethodGet)
	r.HandleFunc(IPNSPath, server.putIPNSRecord).Methods(http.MethodPut)

	return r
}

type server struct {
	svc                   ContentRouter
	disableNDJSON         bool
	recordsLimit          int
	streamingRecordsLimit int
}

func (s *server) provide(w http.ResponseWriter, httpReq *http.Request) {
	req := jsontypes.WriteProvidersRequest{}
	err := json.NewDecoder(httpReq.Body).Decode(&req)
	_ = httpReq.Body.Close()
	if err != nil {
		writeErr(w, "Provide", http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	resp := jsontypes.WriteProvidersResponse{}

	for i, prov := range req.Providers {
		switch v := prov.(type) {
		case *types.WriteBitswapProviderRecord:
			err := v.Verify()
			if err != nil {
				logErr("Provide", "signature verification failed", err)
				writeErr(w, "Provide", http.StatusForbidden, errors.New("signature verification failed"))
				return
			}

			keys := make([]cid.Cid, len(v.Payload.Keys))
			for i, k := range v.Payload.Keys {
				keys[i] = k.Cid
			}
			addrs := make([]multiaddr.Multiaddr, len(v.Payload.Addrs))
			for i, a := range v.Payload.Addrs {
				addrs[i] = a.Multiaddr
			}
			advisoryTTL, err := s.svc.ProvideBitswap(httpReq.Context(), &BitswapWriteProvideRequest{
				Keys:        keys,
				Timestamp:   v.Payload.Timestamp.Time,
				AdvisoryTTL: v.Payload.AdvisoryTTL.Duration,
				ID:          *v.Payload.ID,
				Addrs:       addrs,
			})
			if err != nil {
				writeErr(w, "Provide", http.StatusInternalServerError, fmt.Errorf("delegate error: %w", err))
				return
			}
			resp.ProvideResults = append(resp.ProvideResults,
				&types.WriteBitswapProviderRecordResponse{
					Protocol:    v.Protocol,
					Schema:      v.Schema,
					AdvisoryTTL: &types.Duration{Duration: advisoryTTL},
				},
			)
		case *types.UnknownProviderRecord:
			provResp, err := s.svc.Provide(httpReq.Context(), &WriteProvideRequest{
				Protocol: v.Protocol,
				Schema:   v.Schema,
				Bytes:    v.Bytes,
			})
			if err != nil {
				writeErr(w, "Provide", http.StatusInternalServerError, fmt.Errorf("delegate error: %w", err))
				return
			}
			resp.ProvideResults = append(resp.ProvideResults, provResp)
		default:
			writeErr(w, "Provide", http.StatusBadRequest, fmt.Errorf("provider record %d does not contain a protocol", i))
			return
		}
	}
	writeJSONResult(w, "Provide", resp)
}

func (s *server) findProviders(w http.ResponseWriter, httpReq *http.Request) {
	vars := mux.Vars(httpReq)
	cidStr := vars["cid"]
	cid, err := cid.Decode(cidStr)
	if err != nil {
		writeErr(w, "FindProviders", http.StatusBadRequest, fmt.Errorf("unable to parse CID: %w", err))
		return
	}

	var handlerFunc func(w http.ResponseWriter, provIter iter.ResultIter[types.ProviderResponse])

	var supportsNDJSON bool
	var supportsJSON bool
	var recordsLimit int
	acceptHeaders := httpReq.Header.Values("Accept")
	if len(acceptHeaders) == 0 {
		handlerFunc = s.findProvidersJSON
		recordsLimit = s.recordsLimit
	} else {
		for _, acceptHeader := range acceptHeaders {
			for _, accept := range strings.Split(acceptHeader, ",") {
				mediaType, _, err := mime.ParseMediaType(accept)
				if err != nil {
					writeErr(w, "FindProviders", http.StatusBadRequest, fmt.Errorf("unable to parse Accept header: %w", err))
					return
				}

				switch mediaType {
				case mediaTypeJSON, mediaTypeWildcard:
					supportsJSON = true
				case mediaTypeNDJSON:
					supportsNDJSON = true
				}
			}
		}

		if supportsNDJSON && !s.disableNDJSON {
			handlerFunc = s.findProvidersNDJSON
			recordsLimit = s.streamingRecordsLimit
		} else if supportsJSON {
			handlerFunc = s.findProvidersJSON
			recordsLimit = s.recordsLimit
		} else {
			writeErr(w, "FindProviders", http.StatusBadRequest, errors.New("no supported content types"))
			return
		}
	}

	provIter, err := s.svc.FindProviders(httpReq.Context(), cid, recordsLimit)
	if err != nil {
		writeErr(w, "FindProviders", http.StatusInternalServerError, fmt.Errorf("delegate error: %w", err))
		return
	}

	handlerFunc(w, provIter)
}

func (s *server) findProvidersJSON(w http.ResponseWriter, provIter iter.ResultIter[types.ProviderResponse]) {
	defer provIter.Close()

	var (
		providers []types.ProviderResponse
		i         int
	)

	for provIter.Next() {
		res := provIter.Val()
		if res.Err != nil {
			writeErr(w, "FindProviders", http.StatusInternalServerError, fmt.Errorf("delegate error on result %d: %w", i, res.Err))
			return
		}
		providers = append(providers, res.Val)
		i++
	}
	response := jsontypes.ReadProvidersResponse{Providers: providers}
	writeJSONResult(w, "FindProviders", response)
}

func (s *server) findProvidersNDJSON(w http.ResponseWriter, provIter iter.ResultIter[types.ProviderResponse]) {
	defer provIter.Close()

	w.Header().Set("Content-Type", mediaTypeNDJSON)
	w.WriteHeader(http.StatusOK)
	for provIter.Next() {
		res := provIter.Val()
		if res.Err != nil {
			logger.Errorw("FindProviders ndjson iterator error", "Error", res.Err)
			return
		}
		// don't use an encoder because we can't easily differentiate writer errors from encoding errors
		b, err := drjson.MarshalJSONBytes(res.Val)
		if err != nil {
			logger.Errorw("FindProviders ndjson marshal error", "Error", err)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			logger.Warn("FindProviders ndjson write error", "Error", err)
			return
		}

		_, err = w.Write([]byte{'\n'})
		if err != nil {
			logger.Warn("FindProviders ndjson write error", "Error", err)
			return
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *server) getIPNSRecord(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept"), mediaTypeIPNSRecord) {
		writeErr(w, "GetIPNSRecord", http.StatusNotAcceptable, errors.New("content type in 'Accept' header is missing or not supported"))
		return
	}

	vars := mux.Vars(r)
	cidStr := vars["cid"]
	cid, err := cid.Decode(cidStr)
	if err != nil {
		writeErr(w, "GetIPNSRecord", http.StatusBadRequest, fmt.Errorf("unable to parse CID: %w", err))
		return
	}

	name, err := ipns.NameFromCid(cid)
	if err != nil {
		writeErr(w, "GetIPNSRecord", http.StatusBadRequest, fmt.Errorf("peer ID CID is not valid: %w", err))
		return
	}

	record, err := s.svc.FindIPNSRecord(r.Context(), name)
	if err != nil {
		writeErr(w, "GetIPNSRecord", http.StatusInternalServerError, fmt.Errorf("delegate error: %w", err))
		return
	}

	rawRecord, err := ipns.MarshalRecord(record)
	if err != nil {
		writeErr(w, "GetIPNSRecord", http.StatusInternalServerError, err)
		return
	}

	if ttl, err := record.TTL(); err == nil {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(ttl.Seconds())))
	} else {
		w.Header().Set("Cache-Control", "max-age=60")
	}

	recordEtag := strconv.FormatUint(xxhash.Sum64(rawRecord), 32)
	w.Header().Set("Etag", recordEtag)
	w.Header().Set("Content-Type", mediaTypeIPNSRecord)
	w.Write(rawRecord)
}

func (s *server) putIPNSRecord(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), mediaTypeIPNSRecord) {
		writeErr(w, "PutIPNSRecord", http.StatusNotAcceptable, errors.New("content type in 'Content-Type' header is missing or not supported"))
		return
	}

	vars := mux.Vars(r)
	cidStr := vars["cid"]
	cid, err := cid.Decode(cidStr)
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusBadRequest, fmt.Errorf("unable to parse CID: %w", err))
		return
	}

	name, err := ipns.NameFromCid(cid)
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusBadRequest, fmt.Errorf("peer ID CID is not valid: %w", err))
		return
	}

	// Limit the reader to the maximum record size.
	rawRecord, err := io.ReadAll(io.LimitReader(r.Body, int64(ipns.MaxRecordSize)))
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusBadRequest, fmt.Errorf("provided record is too long: %w", err))
		return
	}

	record, err := ipns.UnmarshalRecord(rawRecord)
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusBadRequest, fmt.Errorf("provided record is invalid: %w", err))
		return
	}

	err = ipns.ValidateWithName(record, name)
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusBadRequest, fmt.Errorf("provided record is invalid: %w", err))
		return
	}

	err = s.svc.ProvideIPNSRecord(r.Context(), name, record)
	if err != nil {
		writeErr(w, "PutIPNSRecord", http.StatusInternalServerError, fmt.Errorf("delegate error: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeJSONResult(w http.ResponseWriter, method string, val any) {
	w.Header().Add("Content-Type", mediaTypeJSON)

	// keep the marshaling separate from the writing, so we can distinguish bugs (which surface as 500)
	// from transient network issues (which surface as transport errors)
	b, err := drjson.MarshalJSONBytes(val)
	if err != nil {
		writeErr(w, method, http.StatusInternalServerError, fmt.Errorf("marshaling response: %w", err))
		return
	}

	_, err = io.Copy(w, bytes.NewBuffer(b))
	if err != nil {
		logErr("Provide", "writing response body", err)
	}
}

func writeErr(w http.ResponseWriter, method string, statusCode int, cause error) {
	w.WriteHeader(statusCode)
	causeStr := cause.Error()
	if len(causeStr) > 1024 {
		causeStr = causeStr[:1024]
	}
	_, err := w.Write([]byte(causeStr))
	if err != nil {
		logErr(method, "error writing error cause", err)
		return
	}
}

func logErr(method, msg string, err error) {
	logger.Infow(msg, "Method", method, "Error", err)
}
