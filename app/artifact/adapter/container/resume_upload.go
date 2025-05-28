// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package container

import (
	"bytes"
	"context"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/auth"
	"github.com/easysoft/gitfox/app/services/protection"
	"github.com/easysoft/gitfox/app/store"
	gitfox_store "github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/types"

	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	_ "crypto/sha256"
)

const (
	logSessionId = "session_id"

	errInvalidToken = "incorrect state"
)

type ResumeRequest struct {
	artStore store.ArtifactStore
	view     *adapter.ViewDescriptor

	digester digest.Digester

	SessionID string
	RepoName  string
	Ref       string

	State *blobUploadState
}

type blobUploadState struct {
	// name is the primary repository under which the blob will be linked.
	Name string

	// UUID identifies the upload.
	SessionID string

	// offset contains the current progress of the upload.
	Offset int64

	// StartedAt is the original start time of the upload.
	StartedAt int64

	// Ref is the storage blob name
	Ref string
}

func (bs *blobUploadState) Pack(key string) (string, error) {
	return hmacKey(key).packUploadState(*bs)
}

type progressStatus string

const (
	progressNotStarted progressStatus = "not_started"
	progressInComplete progressStatus = "incomplete"
	progressCompleted  progressStatus = "completed"
)

type metadataProgress struct {
	Digester string         `json:"digester"`
	Status   progressStatus `json:"status"`
}

// InitResumeRequest generate a random session_id for upload,
// init an empty resume request
func InitResumeRequest(repoName string, artStore store.ArtifactStore, view *adapter.ViewDescriptor) *ResumeRequest {
	newSessionID := uuid.New()
	newRef := strings.ReplaceAll(HashUUID(newSessionID, repoName), "-", "")

	bs := blobUploadState{
		Name:      repoName,
		SessionID: newSessionID.String(),
		Offset:    0,
		StartedAt: time.Now().UnixMilli(),
		Ref:       newRef,
	}

	return &ResumeRequest{
		view:     view,
		artStore: artStore,

		RepoName:  repoName,
		SessionID: bs.SessionID,
		Ref:       bs.Ref,

		State: &bs,
	}
}

func ParseResumeRequest(ctx context.Context, sessionID, fullRepoName, token string, artStore store.ArtifactStore, view *adapter.ViewDescriptor) (*ResumeRequest, error) {
	req := &ResumeRequest{
		artStore: artStore,
		view:     view,
		RepoName: fullRepoName,
	}

	sid, err := uuid.Parse(sessionID)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("invalid session id")
		return nil, ErrBlobUploadInvalid.WithDetail(err.Error())
	}
	req.SessionID = sessionID
	req.Ref = strings.ReplaceAll(HashUUID(sid, fullRepoName), "-", "")

	state, err := hmacKey("").unpackUploadState(token)
	if err != nil {
		return nil, ErrBlobUploadInvalid.WithDetail(err.Error())
	}

	if state.SessionID != req.SessionID {
		log.Ctx(ctx).Info().Msgf("mismatched uuid in upload state: %q != %q", state.SessionID, req.SessionID)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	if state.Ref != req.Ref {
		log.Ctx(ctx).Info().Msgf("mismatched ref in upload state: %q != %q", state.Ref, req.Ref)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	if state.Name != req.RepoName {
		log.Ctx(ctx).Info().Msgf("mismatched repository name in upload state: %q != %q", state.Name, req.RepoName)
		return nil, ErrBlobUploadInvalid.WithDetail(errInvalidToken)
	}

	status, digester, err := parseProgress(ctx, artStore, view, req.Ref)
	if err != nil {
		return nil, err
	}

	if status == progressCompleted {
		return nil, ErrBlobUploadInvalid.WithDetail("upload is already completed")
	}

	req.digester = digester
	req.State = &state
	return req, nil
}

func parseProgress(ctx context.Context, artStore store.ArtifactStore, view *adapter.ViewDescriptor, ref string) (progressStatus, digest.Digester, error) {
	var digester = digest.Canonical.Digester()
	var status progressStatus

	dbBlob, err := artStore.Blobs().GetByRef(ctx, ref, view.StorageID)
	if err != nil {
		if errors.Is(err, gitfox_store.ErrResourceNotFound) {
			status = progressNotStarted
			return status, digester, nil
		} else {
			return "", nil, ErrBlobUploadInvalid.WithDetail(err.Error())
		}
	}

	if dbBlob.Metadata == "" {
		// this should never have happened
		log.Ctx(ctx).Info().Msgf("upload metadata not found in record")
		return "", nil, ErrBlobUploadInvalid.WithDetail("upload metadata not found in server")
	}

	var upMeta metadataProgress
	if err = json.NewDecoder(bytes.NewReader([]byte(dbBlob.Metadata))).Decode(&upMeta); err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("parse upload metadata failed")
	}

	digesterSnapshot, err := base64.StdEncoding.DecodeString(upMeta.Digester)
	if err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resume upload failed")
	}
	rd, ok := digester.Hash().(encoding.BinaryUnmarshaler)
	if !ok {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resumable digest not available")
	}
	if err = rd.UnmarshalBinary(digesterSnapshot); err != nil {
		return "", nil, ErrBlobUploadInvalid.WithDetail("resume upload failed")
	}

	return upMeta.Status, digester, nil
}

func (crr *ResumeRequest) Size(ctx context.Context) (int64, error) {
	fInfo, err := crr.view.Store.Stat(ctx, adapter.BlobPath(crr.Ref))
	if err != nil {
		return 0, err
	}
	return fInfo.Size(), nil
}

func (crr *ResumeRequest) AppendWrite(ctx context.Context, r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != "application/octet-stream" {
		return ErrUnknown.WithDetail("bad Content-Type")
	}

	fw, err := adapter.NewStorageBlobWriter(ctx, crr.view.Store, crr.Ref, true)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("create blob writer failed")
		return ErrBlobUploadUnknown.WithDetail(err.Error())
	}
	defer func() {
		if err != nil {
			log.Ctx(ctx).Debug().Msgf("cancel chunk write on ref %s", crr.Ref)
			fw.Writer.Cancel(ctx)
		} else {
			fw.Writer.Commit(ctx)
		}
		fw.Writer.Close()
	}()

	cr := r.Header.Get("Content-Range")
	cl := r.Header.Get("Content-Length")
	if cr != "" && cl != "" {
		start, end, err := parseContentRange(cr)
		if err != nil {
			return ErrUnknown.WithDetail(err.Error())
		}
		if start > end || start != fw.Writer.Size() {
			return ErrCodeRangeInvalid
		}

		clInt, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return ErrUnknown.WithDetail(err.Error())
		}
		if clInt != (end-start)+1 {
			return ErrSizeInvalid
		}
	}

	log.Ctx(ctx).Debug().Msgf("copy chunk upload")
	if err = copyFullPayload(ctx, r, fw.Writer, crr.digester); err != nil {
		log.Ctx(ctx).Err(err).Msg("copy full payload failed")
		return err
	}

	log.Ctx(ctx).Debug().Msgf("digest after append: %s", crr.digester.Digest())
	crr.State.Offset = fw.Writer.Size()
	if _, err = crr.saveProgress(ctx, progressInComplete, crr.State.Offset); err != nil {
		return err
	}
	return nil
}

func (crr *ResumeRequest) Finish(ctx context.Context, r *http.Request, dgst *digest.Digest) error {
	fw, err := adapter.NewStorageBlobWriter(ctx, crr.view.Store, crr.Ref, true)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("create blob writer failed")
		return ErrBlobUploadUnknown.WithDetail(err.Error())
	}
	defer func() {
		if err != nil {
			log.Ctx(ctx).Err(err).Str(logSessionId, crr.SessionID).Msgf("cancel chunk write on ref %s", crr.Ref)
			fw.Writer.Cancel(ctx)
		} else {
			if e := fw.Writer.Commit(ctx); e != nil {
				log.Ctx(ctx).Err(e).Msgf("commit chunk write on ref %s", crr.Ref)
			}
		}
		fw.Writer.Close()
	}()

	if err = copyFullPayload(ctx, r, fw.Writer, crr.digester); err != nil {
		log.Ctx(ctx).Err(err).Msg("err copy full payload")
		return err
	}
	log.Ctx(ctx).Debug().Msgf("digest after complete: %s", crr.digester.Digest())

	completeDigest := crr.digester.Digest()
	if completeDigest.String() != dgst.String() {
		return ErrBlobUploadInvalid.WithDetail("digest mismatch")
	}

	dbBlob, err := crr.saveProgress(ctx, progressCompleted, fw.Writer.Size())
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("save progress failed")
		return err
	}

	now := time.Now().UnixMilli()
	asset := types.ArtifactAsset{
		Format:      types.ArtifactContainerFormat,
		Path:        dgst.String(),
		ContentType: "application/octet-stream",
		Kind:        types.AssetKindMain,
		BlobID:      dbBlob.ID,
		Created:     now,
		Updated:     now,
	}

	art, err := crr.artStore.Assets().GetPath(ctx, asset.Path, types.ArtifactContainerFormat)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("get artifact by path failed")
		if err = crr.artStore.Assets().Create(ctx, &asset); err != nil {
			log.Ctx(ctx).Err(err).Msg("save artifact asset failed")
			return err
		}
	} else {
		art.BlobID = dbBlob.ID
		art.Updated = now
		if err = crr.artStore.Assets().Update(ctx, art); err != nil {
			log.Ctx(ctx).Err(err).Msg("update artifact asset failed")
			return err
		}
	}

	return nil
}

func (crr *ResumeRequest) Remove(ctx context.Context) error {
	blobPath := adapter.BlobPath(crr.Ref)
	_, err := crr.view.Store.Stat(ctx, blobPath)
	if err == nil {
		log.Debug().Str(logSessionId, crr.SessionID).Msgf("delete blob file %s", blobPath)
		_ = crr.view.Store.Delete(ctx, blobPath)
		return nil
	}
	return nil
}

func (crr *ResumeRequest) saveProgress(ctx context.Context, status progressStatus, size int64) (*types.ArtifactBlob, error) {
	bm, _ := crr.digester.Hash().(encoding.BinaryMarshaler)
	state, err := bm.MarshalBinary()
	if err != nil {
		return nil, ErrUnknown.WithDetail(err.Error())
	}
	b64State := base64.StdEncoding.EncodeToString(state)
	upMeta := metadataProgress{
		Digester: b64State,
		Status:   status,
	}

	mutateFn := func(blob *types.ArtifactBlob) error {
		jMeta, e := protection.ToJSON(&upMeta)
		if e != nil {
			return e
		}
		blob.Metadata = string(jMeta)
		blob.Size = size
		return nil
	}

	var dbBlob *types.ArtifactBlob
	if dbBlob, err = crr.createOrUpdateModelBlob(ctx, mutateFn); err != nil {
		return nil, err
	}

	return dbBlob, nil
}

func (crr *ResumeRequest) createOrUpdateModelBlob(ctx context.Context, mutateFn func(blob *types.ArtifactBlob) error) (*types.ArtifactBlob, error) {
	dbBlob, err := crr.artStore.Blobs().GetByRef(ctx, crr.Ref, crr.view.StorageID)
	if err == nil {
		return crr.artStore.Blobs().UpdateOptLock(ctx, dbBlob, mutateFn)
	}

	if errors.Is(err, gitfox_store.ErrResourceNotFound) {
		var creatorId = auth.AnonymousPrincipal.ID
		session, ok := request.AuthSessionFrom(ctx)
		if ok {
			creatorId = session.User.ID
		}

		newObj := types.ArtifactBlob{
			StorageID: crr.view.StorageID,
			Ref:       crr.Ref,
			Created:   crr.State.StartedAt,
			Creator:   creatorId,
		}
		if err = mutateFn(&newObj); err != nil {
			return nil, err
		}
		if err = crr.artStore.Blobs().Create(ctx, &newObj); err != nil {
			return nil, err
		}
		return &newObj, err
	}
	return nil, err
}

func HashUUID(base uuid.UUID, salt string) string {
	return uuid.NewSHA1(base, []byte(salt)).String()
}

func parseContentRange(cr string) (start int64, end int64, err error) {
	rStart, rEnd, ok := strings.Cut(cr, "-")
	if !ok {
		return -1, -1, fmt.Errorf("invalid content range format, %s", cr)
	}
	start, err = strconv.ParseInt(rStart, 10, 64)
	if err != nil {
		return -1, -1, err
	}
	end, err = strconv.ParseInt(rEnd, 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return start, end, nil
}

// The copy will be limited to `limit` bytes, if limit is greater than zero.
func copyFullPayload(ctx context.Context, r *http.Request, destWriter io.Writer, digester digest.Digester) error {
	// Get a channel that tells us if the client disconnects
	clientClosed := r.Context().Done()

	mw := io.MultiWriter(destWriter, digester.Hash())
	// Read in the data, if any.
	copied, err := io.Copy(mw, r.Body)
	if clientClosed != nil && (err != nil || (r.ContentLength > 0 && copied < r.ContentLength)) {
		// Didn't receive as much content as expected. Did the client
		// disconnect during the request? If so, avoid returning a 400
		// error to keep the logs cleaner.
		select {
		case <-clientClosed:
			return ErrClientClosed.WithDetail("client disconnected")
		default:
		}
	}

	if err != nil {
		log.Ctx(ctx).Err(err).Msg("unknown error reading request payload")
		return ErrUnknown.WithDetail(err.Error())
	}

	return nil
}
