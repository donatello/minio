/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	minio "github.com/minio/minio/cmd"
	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/hash"
	"github.com/valyala/fastjson"
)

var (
	errGWMetaNotFound      = errors.New("dare.meta file not found")
	errGWMetaInvalidFormat = errors.New("dare.meta format is invalid")
)

// A gwMetaV1 represents `gw.json` metadata header.
type gwMetaV1 struct {
	Version string         `json:"version"` // Version of the current `gw.json`.
	Format  string         `json:"format"`  // Format of the current `gw.json`.
	Stat    minio.StatInfo `json:"stat"`    // Stat of the current object `gw.json`.
	ETag    string         `json:"etag"`    // ETag of the current object

	// Metadata map for current object `gw.json`.
	Meta map[string]string `json:"meta,omitempty"`
	// Captures all the individual object `gw.json`.
	Parts []minio.ObjectPartInfo `json:"parts,omitempty"`
}

// Gateway metadata constants.
const (
	// Gateway meta version.
	gwMetaVersion = "1.0.0"

	// Gateway meta version.
	gwMetaVersion100 = "1.0.0"

	// Gateway meta format string.
	gwMetaFormat = "gw"

	// Add new constants here.
)

// newGWMetaV1 - initializes new gwMetaV1, adds version.
func newGWMetaV1() (gwMeta gwMetaV1) {
	gwMeta = gwMetaV1{}
	gwMeta.Version = gwMetaVersion
	gwMeta.Format = gwMetaFormat
	return gwMeta
}

// IsValid - tells if the format is sane by validating the version
// string, format fields.
func (m gwMetaV1) IsValid() bool {
	return ((m.Version == gwMetaVersion || m.Version == gwMetaVersion100) &&
		m.Format == gwMetaFormat)
}

// Converts metadata to object info.
func (m gwMetaV1) ToObjectInfo(bucket, object string) minio.ObjectInfo {
	filterKeys := append([]string{
		"ETag",
		"Content-Length",
		"Last-Modified",
		"Content-Type",
		"Expires",
	}, defaultFilterKeys...)
	objInfo := minio.ObjectInfo{
		IsDir:           false,
		Bucket:          bucket,
		Name:            object,
		Size:            m.Stat.Size,
		ModTime:         m.Stat.ModTime,
		ContentType:     m.Meta["content-type"],
		ContentEncoding: m.Meta["content-encoding"],
		ETag:            minio.CanonicalizeETag(m.ETag),
		UserDefined:     minio.CleanMetadataKeys(m.Meta, filterKeys...),
		Parts:           m.Parts,
	}

	if sc, ok := m.Meta["x-amz-storage-class"]; ok {
		objInfo.StorageClass = sc
	}
	var (
		t time.Time
		e error
	)
	if exp, ok := m.Meta["expires"]; ok {
		if t, e = time.Parse(http.TimeFormat, exp); e == nil {
			objInfo.Expires = t.UTC()
		}
	}
	// Success.
	return objInfo
}

// ObjectToPartOffset - translate offset of an object to offset of its individual part.
func (m gwMetaV1) ObjectToPartOffset(ctx context.Context, offset int64) (partIndex int, partOffset int64, err error) {
	if offset == 0 {
		// Special case - if offset is 0, then partIndex and partOffset are always 0.
		return 0, 0, nil
	}
	partOffset = offset
	// Seek until object offset maps to a particular part offset.
	for i, part := range m.Parts {
		partIndex = i
		// Offset is smaller than size we have reached the proper part offset.
		if partOffset < part.Size {
			return partIndex, partOffset, nil
		}
		// Continue to towards the next part.
		partOffset -= part.Size
	}
	logger.LogIf(ctx, minio.InvalidRange{})
	// Offset beyond the size of the object return InvalidRange.
	return 0, 0, minio.InvalidRange{}
}

// parses gateway metadata stat info from metadata json
func parseGWStat(v *fastjson.Value) (si minio.StatInfo, err error) {
	// obtain stat info.
	st := v.GetObject("stat")
	var mb []byte
	mb, err = st.Get("modTime").StringBytes()
	if err != nil {
		return si, err
	}
	// fetching modTime.
	si.ModTime, err = time.Parse(time.RFC3339, string(mb))
	if err != nil {
		return si, err
	}
	// obtain Stat.Size .
	si.Size, err = st.Get("size").Int64()
	if err != nil {
		return si, err
	}
	return si, nil
}

// parses gateway metadata version from metadata json
func parseGWVersion(v *fastjson.Value) string {
	return string(v.GetStringBytes("version"))
}

// parses gateway ETag from metadata json
func parseGWETag(v *fastjson.Value) string {
	return string(v.GetStringBytes("etag"))
}

// parses gateway metadata format from metadata json
func parseGWFormat(v *fastjson.Value) string {
	return string(v.GetStringBytes("format"))
}

// parses gateway metadata json to get list of ObjectPartInfo
func parseGWParts(v *fastjson.Value) []minio.ObjectPartInfo {
	// Parse the GW Parts.
	partsResult := v.GetArray("parts")
	partInfo := make([]minio.ObjectPartInfo, len(partsResult))
	for i, p := range partsResult {
		partInfo[i] = minio.ObjectPartInfo{
			Number: p.GetInt("number"),
			Name:   string(p.GetStringBytes("name")),
			ETag:   string(p.GetStringBytes("etag")),
			Size:   p.GetInt64("size"),
		}
	}
	return partInfo
}

// parses gateway metadata json to get the metadata map
func parseGWMetaMap(v *fastjson.Value) map[string]string {
	metaMap := make(map[string]string)
	// Get gwMetaV1.Meta map.
	v.GetObject("meta").Visit(func(k []byte, kv *fastjson.Value) {
		metaMap[string(k)] = string(kv.GetStringBytes())
	})
	return metaMap
}

var gwParserPool fastjson.ParserPool

// Constructs GWMetaV1 using `fastjson` lib to retrieve each field.
func gwMetaUnmarshalJSON(ctx context.Context, gwMetaBuf []byte) (gwMeta gwMetaV1, err error) {
	parser := gwParserPool.Get()
	defer gwParserPool.Put(parser)

	var v *fastjson.Value
	v, err = parser.ParseBytes(gwMetaBuf)
	if err != nil {
		return gwMeta, err
	}

	// obtain version.
	gwMeta.Version = parseGWVersion(v)
	// obtain format.
	gwMeta.Format = parseGWFormat(v)
	// Parse gwMetaV1.Stat .
	stat, err := parseGWStat(v)
	if err != nil {
		logger.LogIf(ctx, err)
		return gwMeta, err
	}
	gwMeta.ETag = parseGWETag(v)
	gwMeta.Stat = stat

	// Parse the GW Parts.
	gwMeta.Parts = parseGWParts(v)
	// parse gwMetaV1.
	gwMeta.Meta = parseGWMetaMap(v)

	return gwMeta, nil
}

// readGWMeta reads `dare.meta` and returns back GW metadata structure.
func readGWMetadata(ctx context.Context, buf bytes.Buffer) (gwMeta gwMetaV1, err error) {
	if buf.Len() == 0 {
		return gwMetaV1{}, errGWMetaNotFound
	}
	gwMeta, err = gwMetaUnmarshalJSON(ctx, buf.Bytes())
	if err != nil {
		return gwMetaV1{}, err
	}
	if !gwMeta.IsValid() {
		return gwMetaV1{}, errGWMetaInvalidFormat
	}
	// Return structured `dare.meta`.
	return gwMeta, nil
}

// getGWMetadata - unmarshals dare.meta into a *minio.PutObjReader
func getGWMetadata(ctx context.Context, bucket, prefix string, gwMeta gwMetaV1) (*minio.PutObjReader, error) {
	// Marshal json.
	metadataBytes, err := json.Marshal(&gwMeta)
	if err != nil {
		logger.LogIf(ctx, err)
		return nil, err
	}
	hashReader, err := hash.NewReader(bytes.NewReader(metadataBytes), int64(len(metadataBytes)), "", "", int64(len(metadataBytes)), false)
	if err != nil {
		return nil, err
	}
	return minio.NewPutObjReader(hashReader, nil, nil), nil
}
