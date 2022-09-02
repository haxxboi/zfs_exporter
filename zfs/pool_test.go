package zfs

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestZFSCommandLineParse(t *testing.T) {
	inputStr := `pool: ssd_tank
 state: ONLINE
  scan: scrub repaired 0B in 02:44:52 with 0 errors on Sun Aug 14 03:08:54 2022
config:

        NAME        STATE     READ WRITE CKSUM
        ssd_tank    ONLINE       0    13    26
          mirror-0  ONLINE       1    14    27
            sdc     ONLINE       2    15    28
            sda     ONLINE       3    16    29
          mirror-1  ONLINE       4    17    30
            sdh     ONLINE       5    18    31
            sdd     ONLINE       6    19    32
          mirror-2  ONLINE       7    20    33
            sde     ONLINE       8    21    34
            sdf     ONLINE       9    22    35
          mirror-3  ONLINE      10    23    36
            sdg     ONLINE      11    24    37
            sdi     ONLINE      12    25    38
        spares
          sdj       AVAIL

errors: No known data errors
`
	lines := strings.Split(inputStr, "\n")
	disks, err := parsePoolDisksFromLines(lines)
	if err != nil {
		t.Fatal(err)
	}

	if len(disks) != 13 {
		t.Fatalf("Expected exactly 13 disks output, got %d", len(disks))
	}

	expectedOutput := []PoolDisk{
		{
			Zpool:          "ssd_tank",
			Name:           "mirror-0",
			Vdev:           "mirror-0",
			Kind:           "vdev",
			State:          "ONLINE",
			ReadErrors:     1,
			WriteErrors:    14,
			ChecksumErrors: 27,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-0",
			Name:           "sdc",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     2,
			WriteErrors:    15,
			ChecksumErrors: 28,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-0",
			Name:           "sda",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     3,
			WriteErrors:    16,
			ChecksumErrors: 29,
		},
		{
			Zpool:          "ssd_tank",
			Name:           "mirror-1",
			Vdev:           "mirror-1",
			Kind:           "vdev",
			State:          "ONLINE",
			ReadErrors:     4,
			WriteErrors:    17,
			ChecksumErrors: 30,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-1",
			Name:           "sdh",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     5,
			WriteErrors:    18,
			ChecksumErrors: 31,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-1",
			Name:           "sdd",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     6,
			WriteErrors:    19,
			ChecksumErrors: 32,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-2",
			Name:           "mirror-2",
			Kind:           "vdev",
			State:          "ONLINE",
			ReadErrors:     7,
			WriteErrors:    20,
			ChecksumErrors: 33,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-2",
			Name:           "sde",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     8,
			WriteErrors:    21,
			ChecksumErrors: 34,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-2",
			Name:           "sdf",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     9,
			WriteErrors:    22,
			ChecksumErrors: 35,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-3",
			Name:           "mirror-3",
			Kind:           "vdev",
			State:          "ONLINE",
			ReadErrors:     10,
			WriteErrors:    23,
			ChecksumErrors: 36,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-3",
			Name:           "sdg",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     11,
			WriteErrors:    24,
			ChecksumErrors: 37,
		},
		{
			Zpool:          "ssd_tank",
			Vdev:           "mirror-3",
			Name:           "sdi",
			Kind:           "disk",
			State:          "ONLINE",
			ReadErrors:     12,
			WriteErrors:    25,
			ChecksumErrors: 38,
		},
		{
			Zpool:          "spares",
			Name:           "sdj",
			Kind:           "spare",
			State:          "AVAIL",
			ReadErrors:     0,
			WriteErrors:    0,
			ChecksumErrors: 0,
		},
	}

	diff := cmp.Diff(disks, expectedOutput)
	if diff != "" {
		t.Fatalf("Parsed disks output is not equal to expected output: %s", diff)
	}
}
