package storage

import (
	"fmt"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	cbor "github.com/ipfs/go-ipld-cbor"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/ipfs/go-unixfs/importer/trickle"
	ufsio "github.com/ipfs/go-unixfs/io"
	"github.com/multiformats/go-multihash"
)

func init() {
	ipld.Register(cid.DagProtobuf, dag.DecodeProtobufBlock)
	ipld.Register(cid.Raw, dag.DecodeRawBlock)
	ipld.Register(cid.DagCBOR, cbor.DecodeBlock) // need to decode CBOR
}

type layout = func(db *helpers.DagBuilderHelper) (ipld.Node, error)

const (
	Chunker         string = ""
	DefaultHashFunc        = "sha2-256"
)

// AddStream is suitable for large data
// using trickle layout which is suitable for streaming
func AddStream(node StoragePeer, r io.Reader, hfunc string) (ipld.Node, error) {
	prefix, err := NewCidBuilder(hfunc)
	if err != nil {
		return nil, err
	}

	return Add(node, r, prefix, trickle.Layout)
}

// Add chunks and adds content to the DAGService from a reader.
// data is stored as a UnixFS DAG (default for IPFS).
// fallback to balanced layout, large data should be added via AddStream()
// returns the root ipld.Node
func Add(node StoragePeer, r io.Reader, cb cid.Builder, l layout) (ipld.Node, error) {
	dbp := helpers.DagBuilderParams{
		Dagserv: node.DagService(),
		//RawLeaves:  true,
		Maxlinks: helpers.DefaultLinksPerBlock,
		//NoCopy:     true,
		CidBuilder: cb,
	}

	chnk, err := chunker.FromString(r, Chunker)
	if err != nil {
		return nil, err
	}
	dbh, err := dbp.New(chnk)
	if err != nil {
		return nil, err
	}

	if l == nil {
		l = balanced.Layout
	}

	return l(dbh)
}

// Get returns a reader to a file (must be a UnixFS DAG) as identified by its root CID.
func Get(node StoragePeer, c cid.Cid) (ufsio.ReadSeekCloser, error) {
	dag := node.DagService()
	n, err := dag.Get(node.Context(), c)
	if err != nil {
		return nil, err
	}
	return ufsio.NewDagReader(node.Context(), n, dag)
}

// Get returns a reader to a file (must be a UnixFS DAG) as identified by its root CID.
func GetBytes(node StoragePeer, c cid.Cid) ([]byte, error) {
	rsc, err := Get(node, c)
	defer rsc.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rsc)
}

func NewCidBuilder(hfunc string) (cid.Builder, error) {
	prefix, err := merkledag.PrefixForCidVersion(1)
	if err != nil {
		return nil, fmt.Errorf("bad CID Version: %s", err)
	}
	hashFunCode, ok := multihash.Names[strings.ToLower(hfunc)]
	if !ok {
		hashFunCode = multihash.Names[DefaultHashFunc]
	}
	prefix.MhType = hashFunCode
	prefix.MhLength = -1

	return prefix, nil
}
