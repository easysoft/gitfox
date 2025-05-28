// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package openapi

import (
	"mime/multipart"
	"net/http"

	"github.com/easysoft/gitfox/app/api/request"
	"github.com/easysoft/gitfox/app/api/usererror"
	"github.com/easysoft/gitfox/app/artifact/adapter/container"
	"github.com/easysoft/gitfox/app/artifact/controller"
	"github.com/easysoft/gitfox/types"

	"github.com/gotidy/ptr"
	"github.com/swaggest/openapi-go/openapi3"
)

type listArtifactAssets struct {
	spaceRequest
}

type listArtifactVersions struct {
	spaceRequest
}

type uploadHelmArtifact struct {
	spaceRequest
	Chart multipart.File `formData:"chart"`
}

type uploadRawArtifact struct {
	spaceRequest
	Name    string         `formData:"name"`
	Group   *string        `formData:"group" required:"false"`
	Version string         `formData:"version"`
	File    multipart.File `formData:"file"`
}

type browseArtifactTree struct {
	spaceRequest
	Path   string `formData:"path"`
	Format string `formData:"format"`
}

type removeArtifactNodes struct {
	spaceRequest
	controller.ListNodeInfoRequest
}

var queryParameterQueryArtifactPackage = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamPackage,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact package name"),
		Required:    ptr.Bool(true),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

var queryParameterQueryArtifactGroup = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamGroup,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact package group name, defaults to '', only used for raw and maven format"),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

var queryParameterArtifactPath = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamPath,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact tree node path, exp /a/b/c"),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

var queryParameterArtifactFormat = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryArtifactFormat,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact format"),
		Required:    ptr.Bool(true),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type:    ptrSchemaType(openapi3.SchemaTypeString),
				Default: ptrptr(types.ArtifactAllFormat),
			},
		},
	},
}

var queryParameterArtifactNodeLevel = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamNodeLevel,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact tree node level, asset | version"),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type:    ptrSchemaType(openapi3.SchemaTypeString),
				Default: ptrptr(types.ArtifactNodeLevelAsset),
			},
		},
	},
}

var queryParameterQueryArtifactVersion = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamVersion,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The artifact package version"),
		Required:    ptr.Bool(true),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

var queryParameterQuery = openapi3.ParameterOrRef{
	Parameter: &openapi3.Parameter{
		Name:        request.QueryParamQuery,
		In:          openapi3.ParameterInQuery,
		Description: ptr.String("The query flag"),
		Required:    ptr.Bool(false),
		Schema: &openapi3.SchemaOrRef{
			Schema: &openapi3.Schema{
				Type: ptrSchemaType(openapi3.SchemaTypeString),
			},
		},
	},
}

//nolint:funlen
func artifactOperations(reflector *openapi3.Reflector) {
	opListRawVersions := openapi3.Operation{}
	opListRawVersions.WithTags("artifact")
	opListRawVersions.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactRawVersions"})
	opListRawVersions.WithParameters(queryParameterQueryArtifactPackage, queryParameterQueryArtifactGroup,
		QueryParameterPage, QueryParameterLimit, queryParameterQuery)
	_ = reflector.SetRequest(&opListRawVersions, new(listArtifactVersions), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListRawVersions, []types.ArtifactVersionsRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListRawVersions, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListRawVersions, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListRawVersions, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListRawVersions, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/raw/versions", opListRawVersions)

	opListHelmVersions := openapi3.Operation{}
	opListHelmVersions.WithTags("artifact")
	opListHelmVersions.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactHelmVersions"})
	opListHelmVersions.WithParameters(queryParameterQueryArtifactPackage,
		QueryParameterPage, QueryParameterLimit, queryParameterQuery)
	_ = reflector.SetRequest(&opListHelmVersions, new(listArtifactVersions), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListHelmVersions, []types.ArtifactVersionsRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListHelmVersions, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListHelmVersions, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListHelmVersions, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListHelmVersions, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/helm/versions", opListHelmVersions)

	opListContainerTags := openapi3.Operation{}
	opListContainerTags.WithTags("artifact")
	opListContainerTags.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactContainerTags"})
	opListContainerTags.WithParameters(queryParameterQueryArtifactPackage,
		QueryParameterPage, QueryParameterLimit, queryParameterQuery)
	_ = reflector.SetRequest(&opListContainerTags, new(listArtifactVersions), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListContainerTags, []types.ArtifactVersionsRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListContainerTags, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListContainerTags, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListContainerTags, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListContainerTags, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/container/versions", opListContainerTags)

	opListRawAssets := openapi3.Operation{}
	opListRawAssets.WithTags("artifact")
	opListRawAssets.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactRawAssets"})
	opListRawAssets.WithParameters(queryParameterQueryArtifactPackage, queryParameterQueryArtifactGroup,
		queryParameterQueryArtifactVersion)
	_ = reflector.SetRequest(&opListRawAssets, new(listArtifactAssets), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListRawAssets, []types.ArtifactAssetsRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListRawAssets, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListRawAssets, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListRawAssets, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListRawAssets, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/raw/assets", opListRawAssets)

	opListHelmAssets := openapi3.Operation{}
	opListHelmAssets.WithTags("artifact")
	opListHelmAssets.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactHelmAssets"})
	opListHelmAssets.WithParameters(queryParameterQueryArtifactPackage,
		queryParameterQueryArtifactVersion)
	_ = reflector.SetRequest(&opListHelmAssets, new(listArtifactAssets), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListHelmAssets, []types.ArtifactAssetsRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListHelmAssets, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListHelmAssets, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListHelmAssets, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListHelmAssets, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/helm/assets", opListHelmAssets)

	opListContainerImages := openapi3.Operation{}
	opListContainerImages.WithTags("artifact")
	opListContainerImages.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactContainerImages"})
	opListContainerImages.WithParameters(queryParameterQueryArtifactPackage,
		queryParameterQueryArtifactVersion)
	_ = reflector.SetRequest(&opListContainerImages, new(listArtifactAssets), http.MethodGet)
	_ = reflector.SetJSONResponse(&opListContainerImages, container.TagMetadata{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListContainerImages, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListContainerImages, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListContainerImages, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListContainerImages, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/container/assets", opListContainerImages)

	//
	opBrowseTree := openapi3.Operation{}
	opBrowseTree.WithTags("artifact")
	opBrowseTree.WithMapOfAnything(map[string]interface{}{"operationId": "browseArtifactTree"})
	opBrowseTree.WithParameters(queryParameterArtifactPath, queryParameterArtifactFormat, queryParameterArtifactNodeLevel)
	_ = reflector.SetRequest(&opBrowseTree, new(browseArtifactTree), http.MethodGet)
	_ = reflector.SetJSONResponse(&opBrowseTree, []types.ArtifactTreeRes{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opBrowseTree, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opBrowseTree, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opBrowseTree, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opBrowseTree, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodGet, "/artifacts/{space_ref}/nodes", opBrowseTree)

	opListNodeInfo := openapi3.Operation{}
	opListNodeInfo.WithTags("artifact")
	opListNodeInfo.WithMapOfAnything(map[string]interface{}{"operationId": "listArtifactNodeInfo"})
	_ = reflector.SetRequest(&opListNodeInfo, new(controller.ListNodeInfoRequest), http.MethodPost)
	_ = reflector.SetJSONResponse(&opListNodeInfo, []types.ArtifactNodeInfo{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opListNodeInfo, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opListNodeInfo, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opListNodeInfo, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opListNodeInfo, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodPost, "/artifacts/nodeInfo", opListNodeInfo)

	opRemoveNode := openapi3.Operation{}
	opRemoveNode.WithTags("artifact")
	opRemoveNode.WithMapOfAnything(map[string]interface{}{"operationId": "removeArtifactNode"})
	_ = reflector.SetRequest(&opRemoveNode, new(removeArtifactNodes), http.MethodPost)
	_ = reflector.SetJSONResponse(&opRemoveNode, []types.ArtifactNodeRemoveReport{}, http.StatusOK)
	_ = reflector.SetJSONResponse(&opRemoveNode, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&opRemoveNode, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&opRemoveNode, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&opRemoveNode, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodPost, "/artifacts/{space_ref}/nodes/remove", opRemoveNode)

	opRefreshNodeInfo := openapi3.Operation{}
	opRefreshNodeInfo.WithTags("artifact")
	opRefreshNodeInfo.WithMapOfAnything(map[string]interface{}{"operationId": "refreshArtifactTreeNodes"})
	_ = reflector.SetRequest(&opRefreshNodeInfo, nil, http.MethodPost)
	_ = reflector.SetJSONResponse(&opRefreshNodeInfo, nil, http.StatusOK)
	_ = reflector.SpecEns().AddOperation(http.MethodPost, "/artifacts/refresh_nodes", opRefreshNodeInfo)
}

//nolint:funlen
func artifactUpload(reflector *openapi3.Reflector) {
	helmUpload := openapi3.Operation{}
	helmUpload.WithTags("artifact")
	helmUpload.WithMapOfAnything(map[string]interface{}{"operationId": "uploadHelmPackage"})
	_ = reflector.SetRequest(&helmUpload, new(uploadHelmArtifact), http.MethodPost)
	_ = reflector.SetJSONResponse(&helmUpload, nil, http.StatusCreated)
	_ = reflector.SetJSONResponse(&helmUpload, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&helmUpload, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&helmUpload, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&helmUpload, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodPost, "/artifacts/{space_ref}/helm/upload", helmUpload)

	rawUpload := openapi3.Operation{}
	rawUpload.WithTags("artifact")
	rawUpload.WithMapOfAnything(map[string]interface{}{"operationId": "uploadRawPackage"})
	_ = reflector.SetRequest(&rawUpload, new(uploadRawArtifact), http.MethodPost)
	_ = reflector.SetJSONResponse(&rawUpload, nil, http.StatusCreated)
	_ = reflector.SetJSONResponse(&rawUpload, new(usererror.Error), http.StatusInternalServerError)
	_ = reflector.SetJSONResponse(&rawUpload, new(usererror.Error), http.StatusUnauthorized)
	_ = reflector.SetJSONResponse(&rawUpload, new(usererror.Error), http.StatusForbidden)
	_ = reflector.SetJSONResponse(&rawUpload, new(usererror.Error), http.StatusBadRequest)
	_ = reflector.SpecEns().AddOperation(http.MethodPost, "/artifacts/{space_ref}/raw/upload", rawUpload)
}
