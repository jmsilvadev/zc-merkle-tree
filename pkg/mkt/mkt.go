package mkt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Node represents a node in the Merkle Tree
type Node struct {
	Hash  string
	Left  *Node
	Right *Node
}

// MerkleTree represents the Merkle Tree
type MerkleTree struct {
	Root  *Node
	Nodes []*Node
}

// Proof represents a Merkle proof
type Proof struct {
	Hashes    []string
	Positions []bool
}

// NewMerkleTree creates a new Merkle Tree from a list of hashes
func NewMerkleTree(hashes []string) *MerkleTree {
	var nodes []*Node
	for _, h := range hashes {
		nodes = append(nodes, &Node{Hash: h})
	}

	tree := &MerkleTree{Nodes: nodes}
	tree.Root = buildTree(nodes)
	return tree
}

// GetProof generates a Merkle proof for the given hash
func (mt *MerkleTree) GetProof(hash string) (*Proof, error) {
	var proof Proof
	node := findNode(mt.Root, hash)
	if node == nil {
		return nil, errors.New("hash not found in Merkle tree")
	}

	for node != mt.Root {
		parent := getParent(node, mt.Root)
		if parent == nil {
			return nil, errors.New("parent not found in Merkle tree")
		}

		sibling := getSibling(node, parent)
		if sibling != nil {
			proof.Hashes = append(proof.Hashes, sibling.Hash)
			proof.Positions = append(proof.Positions, parent.Left == node)
		}

		node = parent
	}

	return &proof, nil
}

// PrintAllProofs prints the Merkle proofs for all nodes
// NOTE: I needed this to verify visually the errors during the implementation
// it is not a part of the challenge but helped me
func (mt *MerkleTree) PrintAllProofs() {
	for _, leaf := range mt.Nodes {
		proof, err := mt.GetProof(leaf.Hash)
		if err != nil {
			fmt.Printf("Error generating proof for hash %s: %s\n", leaf.Hash, err)
			continue
		}
		fmt.Printf("Proof for hash %s: %v\n", leaf.Hash, proof)
	}
}

// PrintTree prints the Merkle Tree
// NOTE: I needed this to verify visually the errors during the implementation
// it is not a part of the challenge but helped me
func (mt *MerkleTree) PrintTree() {
	printNode(mt.Root, 0)
}

// VerifyProof verifies a Merkle proof
func VerifyProof(hash, rootHash string, proof *Proof) bool {
	return GetProofHash(hash, proof) == rootHash
}

// GetProofHash returns the rooHash based in the proof given
func GetProofHash(hash string, proof *Proof) string {
	hashStr := hash
	for i, p := range proof.Hashes {
		var combinedHash []byte
		if proof.Positions[i] {
			// If the position is true, the proof hash is on the right
			combinedHash = append([]byte(hashStr), []byte(p)...)
		} else {
			// If the position is false, the proof hash is on the left
			combinedHash = append([]byte(p), []byte(hashStr)...)
		}
		hash := sha256.Sum256(combinedHash)
		hashStr = hex.EncodeToString(hash[:])
	}

	return hashStr
}

// buildTree recursively builds the Merkle Tree
// TODO: change to iterative to save memory and avoid deep recursivity
func buildTree(nodes []*Node) *Node {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var newLevel []*Node
	for i := 0; i < len(nodes); i += 2 {
		if i+1 < len(nodes) {
			hash := sha256.Sum256([]byte(nodes[i].Hash + nodes[i+1].Hash))
			newLevel = append(newLevel, &Node{
				Hash:  hex.EncodeToString(hash[:]),
				Left:  nodes[i],
				Right: nodes[i+1],
			})
		} else {
			newLevel = append(newLevel, nodes[i])
		}
	}

	return buildTree(newLevel)
}

// findNode finds a node with the given hash in the Merkle Tree
// TODO: change to iterative to save memory and avoid deep
// recursivity that can lead to stack overflow
func findNode(root *Node, hash string) *Node {
	if root == nil {
		return nil
	}
	if root.Hash == hash {
		return root
	}
	if node := findNode(root.Left, hash); node != nil {
		return node
	}
	return findNode(root.Right, hash)
}

// getSibling returns the sibling of the given node
func getSibling(node, parent *Node) *Node {
	if parent.Left == node {
		return parent.Right
	}
	return parent.Left
}

// getParent returns the parent of the given node
// TODO: change to iterative to save memory and avoid deep
// recursivity that can lead to stack overflow
func getParent(node, root *Node) *Node {
	if root == nil {
		return nil
	}
	if root.Left == node || root.Right == node {
		return root
	}
	if parent := getParent(node, root.Left); parent != nil {
		return parent
	}
	return getParent(node, root.Right)
}

// printNode prints a node and its children recursively
// TODO: change to iterative to save memory and avoid deep
// recursivity that can lead to stack overflow
func printNode(node *Node, level int) {
	if node == nil {
		return
	}

	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	fmt.Printf("%s%s\n", indent, node.Hash)
	printNode(node.Left, level+1)
	printNode(node.Right, level+1)
}
