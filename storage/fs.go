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
	ipld.Register(cid.DagCBOR, cbor.DecodeBlock)
}

type layout = func(db *helpers.DagBuilderHelper) (ipld.Node, error)

const (
	Chunker         string = ""
	DefaultHashFunc        = "sha2-256"
)

func AddDir(peer StoragePeer) (ufsio.Directory, ipld.Node, error) {
	cb, err := NewCidBuilder("")
	if err != nil {
		return nil, nil, err
	}
	dir := ufsio.NewDirectory(peer.DagService())
	dir.SetCidBuilder(cb)
	dirnode, err := dir.GetNode()
	if err != nil {
		return nil, nil, err
	}
	return dir, dirnode, peer.DagService().Add(peer.Context(), dirnode)
}

func AddToDir(peer StoragePeer, dir ufsio.Directory, name string, node ipld.Node) (ufsio.Directory, ipld.Node, error) {
	err := dir.AddChild(peer.Context(), name, node)
	if err != nil {
		return nil, nil, err
	}
	dirnode, err := dir.GetNode()
	if err != nil {
		return nil, nil, err
	}
	return dir, dirnode, peer.DagService().Add(peer.Context(), dirnode)
}

func LoadDir(peer StoragePeer, c cid.Cid) (ufsio.Directory, error) {
	dag := peer.DagService()
	n, err := dag.Get(peer.Context(), c)
	if err != nil {
		return nil, err
	}
	return ufsio.NewDirectoryFromNode(dag, n)
}

// AddStream is suitable for large data
// using trickle layout which is suitable for streaming
func AddStream(peer StoragePeer, r io.Reader, hfunc string) (ipld.Node, error) {
	cb, err := NewCidBuilder(hfunc)
	if err != nil {
		return nil, err
	}

	return Add(peer, r, cb, trickle.Layout)
}

// Add chunks and adds content to the DAGService from a reader.
// data is stored as a UnixFS DAG (default for IPFS).
// fallback to balanced layout, large data should be added via AddStream()
// returns the root ipld.Node
func Add(peer StoragePeer, r io.Reader, cb cid.Builder, l layout) (ipld.Node, error) {
	dbp := helpers.DagBuilderParams{
		Dagserv: peer.DagService(),
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
func Get(peer StoragePeer, c cid.Cid) (ufsio.ReadSeekCloser, error) {
	dag := peer.DagService()
	n, err := dag.Get(peer.Context(), c)
	if err != nil {
		return nil, err
	}
	return ufsio.NewDagReader(peer.Context(), n, dag)
}

// Get returns a reader to a file (must be a UnixFS DAG) as identified by its root CID.
func GetBytes(peer StoragePeer, c cid.Cid) ([]byte, error) {
	rsc, err := Get(peer, c)
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
