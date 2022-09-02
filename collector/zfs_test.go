package collector

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pdf/zfs_exporter/v2/zfs"
	"github.com/pdf/zfs_exporter/v2/zfs/mock_zfs"
)

func TestZFSCollectInvalidPools(t *testing.T) {
	const result = `# HELP zfs_scrape_collector_duration_seconds zfs_exporter: Duration of a collector scrape.
# TYPE zfs_scrape_collector_duration_seconds gauge
zfs_scrape_collector_duration_seconds{collector="pool"} 0
# HELP zfs_scrape_collector_success zfs_exporter: Whether a collector succeeded.
# TYPE zfs_scrape_collector_success gauge
zfs_scrape_collector_success{collector="pool"} 0
`

	ctrl, ctx := gomock.WithContext(context.Background(), t)
	zfsClient := mock_zfs.NewMockClient(ctrl)
	zfsClient.EXPECT().PoolNames().Return(nil, fmt.Errorf(`Error returned from PoolNames()`)).Times(1)

	config := defaultConfig(zfsClient)
	config.DisableMetrics = false
	collector, err := NewZFS(config)
	collector.Collectors = map[string]State{
		`pool`: {
			Name:       "pool",
			Enabled:    boolPointer(true),
			Properties: stringPointer(``),
			factory:    newPoolCollector,
		},
	}
	if err != nil {
		t.Fatal(err)
	}

	if err = callCollector(ctx, collector, []byte(result), []string{`zfs_scrape_collector_duration_seconds`, `zfs_scrape_collector_success`}); err != nil {
		t.Fatal(err)
	}
}

func TestZFSCollectDisks(t *testing.T) {
	const result = `# HELP zfs_disk_checksum_error zfs_exporter: Disk checksum errors
# TYPE zfs_disk_checksum_error gauge
zfs_disk_checksum_error{disk="mirror-0",kind="vdev",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 27
zfs_disk_checksum_error{disk="sdc",kind="disk",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 28
# HELP zfs_disk_read_error zfs_exporter: Disk read errors
# TYPE zfs_disk_read_error gauge
zfs_disk_read_error{disk="mirror-0",kind="vdev",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 1
zfs_disk_read_error{disk="sdc",kind="disk",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 2
# HELP zfs_disk_status zfs_exporter: Disk status
# TYPE zfs_disk_status gauge
zfs_disk_status{disk="mirror-0",kind="vdev",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 1
zfs_disk_status{disk="sdc",kind="disk",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 1
zfs_disk_status{disk="sdj",kind="spare",state="AVAIL",vdev="",zpool="spares"} 1
# HELP zfs_disk_write_error zfs_exporter: Disk write errors
# TYPE zfs_disk_write_error gauge
zfs_disk_write_error{disk="mirror-0",kind="vdev",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 14
zfs_disk_write_error{disk="sdc",kind="disk",state="ONLINE",vdev="mirror-0",zpool="ssd_tank"} 15
`

	ctrl, ctx := gomock.WithContext(context.Background(), t)
	zfsClient := mock_zfs.NewMockClient(ctrl)
	toReturn := []zfs.PoolDisk{
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
			Zpool:          "spares",
			Name:           "sdj",
			Kind:           "spare",
			State:          "AVAIL",
			ReadErrors:     0,
			WriteErrors:    0,
			ChecksumErrors: 0,
		},
	}
	zfsClient.EXPECT().PoolNames().Return([]string{}, nil)
	zfsClient.EXPECT().PoolDisks().Return(toReturn, nil)

	config := defaultConfig(zfsClient)
	config.DisableMetrics = false
	collector, err := NewZFS(config)
	collector.Collectors = map[string]State{
		`pool-disks`: {
			Name:       "pool-disks",
			Enabled:    boolPointer(true),
			Properties: stringPointer(``),
			factory:    newPoolDiskCollector,
		},
	}
	if err != nil {
		t.Fatal(err)
	}

	expectedNames := []string{
		`zfs_disk_status`,
		`zfs_disk_read_error`,
		`zfs_disk_write_error`,
		`zfs_disk_checksum_error`,
	}
	if err = callCollector(ctx, collector, []byte(result), expectedNames); err != nil {
		t.Fatal(err)
	}
}
