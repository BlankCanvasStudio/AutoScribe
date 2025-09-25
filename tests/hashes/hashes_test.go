package hashdb

import (
    "bytes"
    "crypto/rand"
    "os"
    "sort"
    "testing"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)

func TestCheckFileForHash(t *testing.T) {
	// Generate deterministic-ish data
	makeHash := func(b byte) [32]byte {
		var h [32]byte
		for i := range h {
			h[i] = b
		}
		return h
	}

	// Base set (sorted by construction)
	var hashes [][32]byte
	for i := 10; i < 20; i++ {
		hashes = append(hashes, makeHash(byte(i)))
	}

	// Add some randoms, then sort to ensure file is lexicographically ordered
	for i := 0; i < 10; i++ {
		var h [32]byte
		if _, err := rand.Read(h[:]); err != nil {
			t.Fatalf("rand: %v", err)
		}
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i][:], hashes[j][:]) < 0
	})

	// Write file: each record is 32 bytes + '\n'
	f, err := os.CreateTemp(t.TempDir(), "hashdb-*.dat")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	filename := f.Name()
	for _, h := range hashes {
		if _, err := f.Write(h[:]); err != nil {
			t.Fatalf("write hash: %v", err)
		}
		if _, err := f.Write([]byte{'\n'}); err != nil {
			t.Fatalf("write newline: %v", err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Pick an existing hash and a non-existing one
	existing := hashes[len(hashes)/2]
	absent := makeHash(0xFF)
	// Ensure absent not accidentally in set
	idx := sort.Search(len(hashes), func(i int) bool {
		return bytes.Compare(hashes[i][:], absent[:]) >= 0
	})
	if idx < len(hashes) && bytes.Equal(hashes[idx][:], absent[:]) {
		t.Skip("randomized data happened to include the 'absent' sentinel; skip")
	}

	ok, _, err := mst.CheckFileForHash(filename, existing)
	if err != nil {
		t.Fatalf("mst.CheckFileForHash(existing) error: %v", err)
	}
	if !ok {
		t.Fatalf("expected to find existing hash")
	}

	ok, _, err = mst.CheckFileForHash(filename, absent)
	if err != nil {
		t.Fatalf("mst.CheckFileForHash(absent) error: %v", err)
	}
	if ok {
		t.Fatalf("did not expect to find absent hash")
	}
}

func readAllHashes(t *testing.T, filename string) [][32]byte {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	const rec = 33 // 32 bytes + '\n'
	if len(data)%rec != 0 {
		t.Fatalf("file length %d not multiple of %d", len(data), rec)
	}
	n := len(data) / rec
	out := make([][32]byte, 0, n)
	for i := 0; i < n; i++ {
		var h [32]byte
		copy(h[:], data[i*rec:i*rec+32])
		out = append(out, h)
	}
	return out
}

func makeHash(b byte) [32]byte {
	var h [32]byte
	for i := range h {
		h[i] = b
	}
	return h
}

func writeHashes(t *testing.T, filename string, hashes [][32]byte) {
	t.Helper()
	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, h := range hashes {
		if _, err := f.Write(h[:]); err != nil {
			t.Fatalf("write hash: %v", err)
		}
		if _, err := f.Write([]byte{'\n'}); err != nil {
			t.Fatalf("write newline: %v", err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func assertSorted(t *testing.T, hashes [][32]byte) {
	t.Helper()
	if !sort.SliceIsSorted(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i][:], hashes[j][:]) < 0
	}) {
		t.Fatalf("hash list is not sorted")
	}
}

func TestInsertHash(t *testing.T) {
	// Build a sorted base set
	var base [][32]byte
	for i := 10; i < 20; i++ {
		base = append(base, makeHash(byte(i)))
	}
	// Add randoms, then sort to ensure file order
	for i := 0; i < 10; i++ {
		var h [32]byte
		if _, err := rand.Read(h[:]); err != nil {
			t.Fatalf("rand: %v", err)
		}
		base = append(base, h)
	}
	sort.Slice(base, func(i, j int) bool { return bytes.Compare(base[i][:], base[j][:]) < 0 })
	// Create temp file with initial data
	tmp := t.TempDir()
	filename := tmp + "/hashdb.dat"
	writeHashes(t, filename, base)
	
	// Case 1: insert at front
	front := [32]byte{}
	copy(front[:], bytes.Repeat([]byte{0x00}, 32))
	if bytes.Compare(front[:], base[0][:]) >= 0 {
		front = makeHash(base[0][0] - 1) // ensure smaller than first
	}
	pos, err := mst.InsertHash(filename, front)
	if err != nil {
		t.Fatalf("InsertHash(front): %v", err)
	}
	if pos != 0 {
		t.Fatalf("InsertHash(front): expected position 0, got %d", pos)
	}
	all := readAllHashes(t, filename)
	assertSorted(t, all)
	if ok, _, err := mst.CheckFileForHash(filename, front); err != nil || !ok {
		t.Fatalf("front not found or error: ok=%v err=%v", ok, err)
	}
	
	// Case 2: insert in middle
	midIdx := len(all) / 2
	midLo := all[midIdx-1]
	midHi := all[midIdx]
	// choose a value between midLo and midHi (tweak last byte)
	mid := midLo
	if bytes.Equal(midLo[:], midHi[:]) || bytes.Compare(midLo[:], midHi[:]) >= 0 {
		t.Fatalf("unexpected ordering around middle")
	}
	mid[31] = (midLo[31] + midHi[31]) / 2
	// ensure strictly between; if collision, bump
	if bytes.Compare(mid[:], midLo[:]) <= 0 {
		mid[31] = midLo[31] + 1
	}
	if bytes.Compare(mid[:], midHi[:]) >= 0 {
		mid[31] = midHi[31] - 1
	}
	pos, err = mst.InsertHash(filename, mid)
	if err != nil {
		t.Fatalf("InsertHash(middle): %v", err)
	}
	if pos != int64(midIdx) {
		t.Fatalf("InsertHash(middle): expected position %d, got %d", midIdx, pos)
	}
	all = readAllHashes(t, filename)
	assertSorted(t, all)
	if ok, _, err := mst.CheckFileForHash(filename, mid); err != nil || !ok {
		t.Fatalf("middle not found or error: ok=%v err=%v", ok, err)
	}
	
	// Case 3: insert at end
	end := makeHash(0xFF)
	// ensure strictly greater than current last
	if bytes.Compare(end[:], all[len(all)-1][:]) <= 0 {
		end[0] = 0xFF
		end[31] = 0xFF
	}
	expectedEndPos := int64(len(all))
	pos, err = mst.InsertHash(filename, end)
	if err != nil {
		t.Fatalf("InsertHash(end): %v", err)
	}
	if pos != expectedEndPos {
		t.Fatalf("InsertHash(end): expected position %d, got %d", expectedEndPos, pos)
	}
	all = readAllHashes(t, filename)
	assertSorted(t, all)
	if ok, _, err := mst.CheckFileForHash(filename, end); err != nil || !ok {
		t.Fatalf("end not found or error: ok=%v err=%v", ok, err)
	}
	
	// Case 4: insert duplicate (should be idempotent)
	before := readAllHashes(t, filename)
	duplicateIdx := len(before) / 3
	duplicate := before[duplicateIdx]
	pos, err = mst.InsertHash(filename, duplicate)
	if err != nil {
		t.Fatalf("InsertHash(duplicate): %v", err)
	}
	if pos != int64(duplicateIdx) {
		t.Fatalf("InsertHash(duplicate): expected position %d, got %d", duplicateIdx, pos)
	}
	after := readAllHashes(t, filename)
	if len(after) != len(before) {
		t.Fatalf("duplicate insert changed length: before=%d after=%d", len(before), len(after))
	}
	assertSorted(t, after)
}

func TestInsertHash_CreateIfMissing(t *testing.T) {
	filename := t.TempDir() + "/new-hashdb.dat"
	h := makeHash(0x42)
	if _, err := mst.InsertHash(filename, h); err != nil {
		t.Fatalf("InsertHash(create): %v", err)
	}
	all := readAllHashes(t, filename)
	if len(all) != 1 || !bytes.Equal(all[0][:], h[:]) {
		t.Fatalf("unexpected contents after create: %d entries", len(all))
	}
	if ok, _, err := mst.CheckFileForHash(filename, h); err != nil || !ok {
		t.Fatalf("created hash not found or error: ok=%v err=%v", ok, err)
	}
}

