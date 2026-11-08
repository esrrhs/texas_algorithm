// Command packcolv turns the fixed-width master tables produced by genrate
// into the runtime TXV1 columnar + zstd files:
//
//	go run ./cmd/packcolv -dir ../..
//
// Master layout (master_winrate_N.bin, 16 bytes per row):
//
//	magic "TXM1"  n u8  reserved[3]  rows u64  total u64
//	rows: key u64  win u32  tie u32
//
// Runtime layout (texas_data_winrate_N.zst, zstd-compressed payload):
//
//	"TXV1" (4) | n (1) | reserved (3) | rows (u64 LE) | total (u64 LE)
//	rows x uvarint   key delta against the previous row
//	rows x uvarint   win count
//	rows x uvarint   tie count
//
// total is derived from the orbit size, so it is stored once per level; loss
// is never written (loss = total - win - tie). Keys are emitted in ascending
// order and delta-encoded, which is where most of the compression comes from.
//
// The master is scanned once per column instead of being held in memory, so
// level 5 (123M rows, about 2 GB of master) packs in a flat heap.
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

func main() {
	dir := flag.String("dir", ".", "directory holding master_winrate_N.bin")
	from := flag.Int("from", 0, "first level to pack")
	to := flag.Int("to", 5, "last level to pack")
	flag.Parse()

	for n := *from; n <= *to; n++ {
		if err := packLevel(*dir, n); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func packLevel(dir string, n int) error {
	master := filepath.Join(dir, fmt.Sprintf("master_winrate_%d.bin", n))
	out := filepath.Join(dir, fmt.Sprintf("texas_data_winrate_%d.zst", n))

	f, err := os.Open(master)
	if err != nil {
		return err
	}
	br := bufio.NewReaderSize(f, 1<<20)
	var hdr [24]byte
	if _, err := io.ReadFull(br, hdr[:]); err != nil {
		f.Close()
		return err
	}
	if string(hdr[0:4]) != "TXM1" {
		f.Close()
		return fmt.Errorf("%s: bad master magic %q", master, string(hdr[0:4]))
	}
	if int(hdr[4]) != n {
		f.Close()
		return fmt.Errorf("%s: master holds level %d", master, hdr[4])
	}
	rows := binary.LittleEndian.Uint64(hdr[8:16])
	total := binary.LittleEndian.Uint64(hdr[16:24])
	if rows == 0 {
		f.Close()
		return fmt.Errorf("%s: no rows", master)
	}

	tmp := out + ".tmp"
	w, err := os.Create(tmp)
	if err != nil {
		f.Close()
		return err
	}
	enc, err := zstd.NewWriter(w,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression),
		zstd.WithEncoderConcurrency(1),
	)
	if err != nil {
		w.Close()
		f.Close()
		return err
	}
	bw := bufio.NewWriterSize(enc, 1<<20)

	// Header: total is a theoretical constant per level, not a statistic.
	var rhdr [24]byte
	copy(rhdr[0:4], "TXV1")
	rhdr[4] = byte(n)
	binary.LittleEndian.PutUint64(rhdr[8:16], rows)
	binary.LittleEndian.PutUint64(rhdr[16:24], total)
	if _, err := bw.Write(rhdr[:]); err != nil {
		return err
	}

	// Three column passes. Each re-reads the fixed-width master, which keeps the
	// heap flat: only one 16-byte record is alive at a time.
	var rec [16]byte
	var buf [16]byte
	for _, col := range []int{0, 1, 2} {
		if _, err := f.Seek(24, io.SeekStart); err != nil {
			return err
		}
		br.Reset(f)
		var prev uint64
		for i := uint64(0); i < rows; i++ {
			if _, err := io.ReadFull(br, rec[:]); err != nil {
				return fmt.Errorf("%s: row %d: %w", master, i, err)
			}
			switch col {
			case 0:
				key := binary.LittleEndian.Uint64(rec[0:8])
				if i > 0 && key <= prev {
					return fmt.Errorf("%s: keys are not strictly increasing at row %d", master, i)
				}
				p := binary.PutUvarint(buf[:], key-prev)
				if _, err := bw.Write(buf[:p]); err != nil {
					return err
				}
				prev = key
			case 1:
				p := binary.PutUvarint(buf[:], uint64(binary.LittleEndian.Uint32(rec[8:12])))
				if _, err := bw.Write(buf[:p]); err != nil {
					return err
				}
			case 2:
				p := binary.PutUvarint(buf[:], uint64(binary.LittleEndian.Uint32(rec[12:16])))
				if _, err := bw.Write(buf[:p]); err != nil {
					return err
				}
			}
		}
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	f.Close()
	if err := os.Rename(tmp, out); err != nil {
		os.Remove(tmp)
		return err
	}
	if fi, err := os.Stat(out); err == nil {
		fmt.Printf("wrote %s rows=%d total=%d bytes=%d\n", out, rows, total, fi.Size())
	}
	return nil
}
