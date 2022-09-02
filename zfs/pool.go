package zfs

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
)

// PoolStatus enum contains status text
type PoolStatus string

const (
	// PoolOnline enum entry
	PoolOnline PoolStatus = `ONLINE`
	// PoolDegraded enum entry
	PoolDegraded PoolStatus = `DEGRADED`
	// PoolFaulted enum entry
	PoolFaulted PoolStatus = `FAULTED`
	// PoolOffline enum entry
	PoolOffline PoolStatus = `OFFLINE`
	// PoolUnavail enum entry
	PoolUnavail PoolStatus = `UNAVAIL`
	// PoolRemoved enum entry
	PoolRemoved PoolStatus = `REMOVED`
	// PoolSuspended enum entry
	PoolSuspended PoolStatus = `SUSPENDED`
)

type poolImpl struct {
	name string
}

func (p poolImpl) Name() string {
	return p.name
}

func (p poolImpl) Properties(props ...string) (PoolProperties, error) {
	handler := newPoolPropertiesImpl()
	if err := execute(p.name, handler, `zpool`, `get`, `-Hpo`, `name,property,value`, strings.Join(props, `,`)); err != nil {
		return handler, err
	}
	return handler, nil
}

type poolPropertiesImpl struct {
	properties map[string]string
}

func (p *poolPropertiesImpl) Properties() map[string]string {
	return p.properties
}

// processLine implements the handler interface
func (p *poolPropertiesImpl) processLine(pool string, line []string) error {
	if len(line) != 3 || line[0] != pool {
		return ErrInvalidOutput
	}
	p.properties[line[1]] = line[2]

	return nil
}

// PoolNames returns a list of available pool names
func poolNames() ([]string, error) {
	pools := make([]string, 0)
	cmd := exec.Command(`zpool`, `list`, `-Ho`, `name`)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		pools = append(pools, scanner.Text())
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return pools, nil
}

func newPoolImpl(name string) poolImpl {
	return poolImpl{
		name: name,
	}
}

func newPoolPropertiesImpl() *poolPropertiesImpl {
	return &poolPropertiesImpl{
		properties: make(map[string]string),
	}
}

// Example string to parse:
//
// pool: ssd_tank
//  state: ONLINE
//   scan: scrub repaired 0B in 02:44:52 with 0 errors on Sun Aug 14 03:08:54 2022
// config:

//         NAME        STATE     READ WRITE CKSUM
//         ssd_tank    ONLINE       0     0     0
//           mirror-0  ONLINE       0     0     0
//             sdc     ONLINE       0     0     0
//             sda     ONLINE       0     0     0
//           mirror-1  ONLINE       0     0     0
//             sdh     ONLINE       0     0     0
//             sdd     ONLINE       0     0     0
//           mirror-2  ONLINE       0     0     0
//             sde     ONLINE       0     0     0
//             sdf     ONLINE       0     0     0
//           mirror-3  ONLINE       0     0     0
//             sdg     ONLINE       0     0     0
//             sdi     ONLINE       0     0     0
//         spares
//           sdj       AVAIL

// errors: No known data errors
func poolDisks() ([]PoolDisk, error) {
	lines := make([]string, 0)
	cmd := exec.Command(`zpool`, `status`, `-L`)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(out)

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := strings.ReplaceAll(scanner.Text(), "\t", "        ")
		lines = append(lines, line)
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return parsePoolDisksFromLines(lines)
}

func parsePoolDisksFromLines(lines []string) ([]PoolDisk, error) {
	// little more than we need but not by much
	poolDisks := make([]PoolDisk, 0, len(lines))
	isInsideDisks := false
	minPadding := 0
	currentZpool := ""
	currentVdev := ""
	for _, line := range lines {
		if !isInsideDisks {
			if strings.Contains(line, "NAME") && strings.Contains(line, "STATE") && strings.Contains(line, "CKSUM") {
				isInsideDisks = true
				minPadding = len(line) - len(strings.TrimLeft(line, " "))
				continue
			}
		} else {
			currentPadding := len(line) - len(strings.TrimLeft(line, " "))
			if currentPadding >= minPadding {
				fields := strings.Fields(line)
				if currentPadding-minPadding == 0 {
					// zpool level
					if len(fields) > 0 {
						currentZpool = fields[0]
					}
				} else if currentPadding-minPadding == 2 {
					if currentZpool == "spares" {
						// spares
						if len(fields) == 2 {
							poolDisks = append(poolDisks, PoolDisk{
								Zpool: "spares",
								Name:  fields[0],
								Kind:  "spare",
								State: fields[1],
							})
						}
					} else {
						// vdevs
						if len(fields) == 5 {
							currentVdev = fields[0]
							readErrors, err := strconv.Atoi(fields[2])
							if err != nil {
								return nil, err
							}
							writeErrors, err := strconv.Atoi(fields[3])
							if err != nil {
								return nil, err
							}
							checksumErrors, err := strconv.Atoi(fields[4])
							if err != nil {
								return nil, err
							}

							poolDisks = append(poolDisks, PoolDisk{
								Zpool:          currentZpool,
								Name:           currentVdev,
								Vdev:           currentVdev,
								Kind:           "vdev",
								State:          fields[1],
								ReadErrors:     readErrors,
								WriteErrors:    writeErrors,
								ChecksumErrors: checksumErrors,
							})
						}
					}
				} else if currentPadding-minPadding >= 4 {
					// physical device level
					if len(fields) == 5 {
						readErrors, err := strconv.Atoi(fields[2])
						if err != nil {
							return nil, err
						}
						writeErrors, err := strconv.Atoi(fields[3])
						if err != nil {
							return nil, err
						}
						checksumErrors, err := strconv.Atoi(fields[4])
						if err != nil {
							return nil, err
						}

						poolDisks = append(poolDisks, PoolDisk{
							Zpool:          currentZpool,
							Vdev:           currentVdev,
							Name:           fields[0],
							Kind:           "disk",
							State:          fields[1],
							ReadErrors:     readErrors,
							WriteErrors:    writeErrors,
							ChecksumErrors: checksumErrors,
						})
					}
				}
			} else {
				break
			}
		}
	}

	return poolDisks, nil
}
