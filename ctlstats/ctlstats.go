// SPDX-Licence-Identifier: Apache-2.0 AND BSD-2-Clause-FreeBSD
package ctlstats

/*
#cgo CFLAGS: -I/usr/src/sys
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/param.h>
#include <sys/time.h>
#include <sys/sysctl.h>
#include <sys/resource.h>
#include <sys/queue.h>
#include <sys/callout.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <getopt.h>
#include <string.h>
#include <errno.h>
#include <err.h>
#include <ctype.h>
#include <bitstring.h>
#include <cam/scsi/scsi_all.h>
#include <cam/ctl/ctl.h>
#include <cam/ctl/ctl_io.h>
#include <cam/ctl/ctl_scsi_all.h>
#include <cam/ctl/ctl_util.h>
#include <cam/ctl/ctl_backend.h>
#include <cam/ctl/ctl_ioctl.h>
#include <stdint.h>
#include <sys/types.h>
#include <sys/times.h>
#include <cam/ctl/ctl_ioctl.h>
#include <cam/ctl/ctl.h>
#include <stdio.h>

#define	CTL_STAT_NUM_ITEMS	256

int getstats(int fd, int *alloc_items, int *num_items, struct ctl_io_stats **stats)
{
	// The code of this function comes from ctlstat.c of the freeBSD project
	// It is covered by BSD-2-Clause-FreeBSD, see original file for more details
	// Copyright (c) 2004, 2008, 2009 Silicon Graphics International Corp.
	// Copyright (c) 2017 Alexander Motin <mav@FreeBSD.org>
	// All rights reserved.
	struct ctl_get_io_stats get_stats;
	int more_space_count = 0;

	if (*alloc_items == 0)
		*alloc_items = CTL_STAT_NUM_ITEMS;
retry:
	if (*stats == NULL)
		*stats = malloc(sizeof(**stats) * *alloc_items);

	memset(&get_stats, 0, sizeof(get_stats));
	get_stats.alloc_len = *alloc_items * sizeof(**stats);
	memset(*stats, 0, get_stats.alloc_len);
	get_stats.stats = *stats;

	if (ioctl(fd, CTL_GET_LUN_STATS, &get_stats) == -1)
		err(1, "CTL_GET_*_STATS ioctl returned error");

	switch (get_stats.status) {
	case CTL_SS_OK:
		break;
	case CTL_SS_ERROR:
		err(1, "CTL_GET_*_STATS ioctl returned CTL_SS_ERROR");
		break;
	case CTL_SS_NEED_MORE_SPACE:
		if (more_space_count >= 2)
			errx(1, "CTL_GET_*_STATS returned NEED_MORE_SPACE again");
		*alloc_items = get_stats.num_items * 5 / 4;
		free(*stats);
		*stats = NULL;
		more_space_count++;
		goto retry;
		break;
	default:
		errx(1, "CTL_GET_*_STATS ioctl returned unknown status %d", get_stats.status);
		break;
	}

	*num_items = get_stats.fill_len / sizeof(**stats);
	return (0);
}

char* getports(int fd, int* port_len){
	// The code of this function comes from ctladm.c of the FreeBSD project
	// It is covered by BSD-2-Clause-FreeBSD
	// Copyright (c) 2003, 2004 Silicon Graphics International Corp.
	// Copyright (c) 1997-2007 Kenneth D. Merry
	// Copyright (c) 2012 The FreeBSD Foundation
	// Copyright (c) 2018 Marcelo Araujo <araujo@FreeBSD.org>
	// All rights reserved.

	struct ctl_lun_list list;
	char* port_str = NULL;
retry:
	port_str = (char *)realloc(port_str, *port_len);

	bzero(&list, sizeof(list));
	list.alloc_len = *port_len;
	list.status = CTL_LUN_LIST_NONE;
	list.lun_xml = port_str;

	if (ioctl(fd, CTL_PORT_LIST, &list) == -1) {
		warn("%s: error issuing CTL_PORT_LIST ioctl", __func__);
	}

	if (list.status == CTL_LUN_LIST_ERROR) {
		warnx("%s: error returned from CTL_PORT_LIST ioctl:\n%s",
			__func__, list.error_str);
	} else if (list.status == CTL_LUN_LIST_NEED_MORE_SPACE) {
		*port_len = *port_len << 1;
		goto retry;
	}
	return port_str;
}
*/
import "C"
import (
	"encoding/xml"
	"log"
	"os"
	"unsafe"
)

type Ctl_stats_types int

const (
	CTL_STATS_NO_IO Ctl_stats_types = iota
	CTL_STATS_READ
	CTL_STATS_WRITE
)

// Statistics exported for a specific LUN or Port
type Ctl_io_stats struct {
	Bytes      map[Ctl_stats_types]uint64
	Operations map[Ctl_stats_types]uint64
	Dmas       map[Ctl_stats_types]uint64
}

// Wrapper around CGO code to collect the statistics of every LUN
func GetStats() map[uint32]Ctl_io_stats {
	fd, _ := os.Open(C.CTL_DEFAULT_DEV)
	var alloc_items C.int
	var num_items C.int
	var stats *C.struct_ctl_io_stats
	output := map[uint32]Ctl_io_stats{}
	C.getstats(C.int(fd.Fd()), &alloc_items, &num_items, &stats)
	io_stats := (*[1 << 28]C.struct_ctl_io_stats)(unsafe.Pointer(stats))[:num_items:alloc_items]
	for _, item := range io_stats {
		output[uint32(item.item)] = Ctl_io_stats{
			Bytes: map[Ctl_stats_types]uint64{
				CTL_STATS_NO_IO: uint64(item.bytes[CTL_STATS_NO_IO]),
				CTL_STATS_READ:  uint64(item.bytes[CTL_STATS_READ]),
				CTL_STATS_WRITE: uint64(item.bytes[CTL_STATS_WRITE]),
			},
			Operations: map[Ctl_stats_types]uint64{
				CTL_STATS_NO_IO: uint64(item.operations[CTL_STATS_NO_IO]),
				CTL_STATS_READ:  uint64(item.operations[CTL_STATS_READ]),
				CTL_STATS_WRITE: uint64(item.operations[CTL_STATS_WRITE]),
			},
			Dmas: map[Ctl_stats_types]uint64{
				CTL_STATS_NO_IO: uint64(item.dmas[CTL_STATS_NO_IO]),
				CTL_STATS_READ:  uint64(item.dmas[CTL_STATS_READ]),
				CTL_STATS_WRITE: uint64(item.dmas[CTL_STATS_WRITE]),
			},
		}
	}
	return output
}

type LUN struct {
	Id        uint `xml:"id,attr"`
	LunNumber uint `xml:",chardata"`
}

type Initiator struct {
	Id   uint   `xml:"id,attr"`
	Name string `xml:",chardata"`
}

type Target struct {
	Name       string      `xml:"target"`
	LUN        []LUN       `xml:"lun"`
	Initiators []Initiator `xml:"initiator"`
}

type CtlPortList struct {
	Targets  []Target `xml:"targ_port"`
	lunTable map[uint]uint
}

func (cp *CtlPortList) populateLunTable() {
	cp.lunTable = make(map[uint]uint)
	for index, target := range cp.Targets {
		for _, lun := range target.LUN {
			cp.lunTable[lun.LunNumber] = uint(index)
		}
	}
}

// Do a reverse match of Kernel Lun number to Target
func (cp CtlPortList) GetLunTarget(lunNumber uint) *Target {
	return &cp.Targets[cp.lunTable[lunNumber]]
}

// Get Lun ID in target from kernel LUN number
func (cp CtlPortList) GetLunId(lunNumber uint) uint {
	for _, lun := range cp.GetLunTarget(lunNumber).LUN {
		if lun.LunNumber == lunNumber {
			return lun.Id
		}
	}
	return 0
}

// Wrapper around CGO code to get the list of targets
func GetTargets() CtlPortList {
	fd, _ := os.Open(C.CTL_DEFAULT_DEV)
	output := CtlPortList{}
	var xmlLen C.int = 4096
	xmlString := C.GoStringN(C.getports(C.int(fd.Fd()), &xmlLen), xmlLen)
	err := xml.Unmarshal([]byte(xmlString), &output)
	if err != nil {
		log.Println(err)
	}
	output.populateLunTable()
	return output
}
