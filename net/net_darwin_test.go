package net

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

const (
	netstatTruncated = `Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll Drop
lo0   16384 <Link#1>                         31241     0    3769823    31241     0    3769823     0   0
lo0   16384 ::1/128     ::1                  31241     -    3769823    31241     -    3769823     -   -
lo0   16384 127           127.0.0.1          31241     -    3769823    31241     -    3769823     -   -
lo0   16384 fe80::1%lo0 fe80:1::1            31241     -    3769823    31241     -    3769823     -   -
gif0* 1280  <Link#2>                             0     0          0        0     0          0     0   0
stf0* 1280  <Link#3>                             0     0          0        0     0          0     0   0
utun8 1500  <Link#88>                          286     0      27175        0     0          0     0   0
utun8 1500  <Link#90>                          286     0      29554        0     0          0     0   0
utun8 1500  <Link#92>                          286     0      29244        0     0          0     0   0
utun8 1500  <Link#93>                          286     0      28267        0     0          0     0   0
utun8 1500  <Link#95>                          286     0      28593        0     0          0     0   0`
	netstatNotTruncated = `Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll Drop
lo0   16384 <Link#1>                      27190978     0 12824763793 27190978     0 12824763793     0   0
lo0   16384 ::1/128     ::1               27190978     - 12824763793 27190978     - 12824763793     -   -
lo0   16384 127           127.0.0.1       27190978     - 12824763793 27190978     - 12824763793     -   -
lo0   16384 fe80::1%lo0 fe80:1::1         27190978     - 12824763793 27190978     - 12824763793     -   -
gif0* 1280  <Link#2>                             0     0          0        0     0          0     0   0
stf0* 1280  <Link#3>                             0     0          0        0     0          0     0   0
en0   1500  <Link#4>    a8:66:7f:dd:ee:ff  5708989     0 7295722068  3494252     0  379533492     0 230
en0   1500  fe80::aa66: fe80:4::aa66:7fff  5708989     - 7295722068  3494252     -  379533492     -   -`
	ifconfigOutput = `lo0: flags=8049<UP,LOOPBACK,RUNNING,MULTICAST> mtu 16384
	options=1203<RXCSUM,TXCSUM,TXSTATUS,SW_TIMESTAMP>
	inet 192.168.0.100 netmask 255.255.255.0
	inet6 fe80::1234:5678:abcd:ef01%lo0 prefixlen 64 scopeid 0x1
	inet6 fe80::5678:abcd:ef01:1234%lo0 prefixlen 64 scopeid 0x1
	nd6 options=201<PERFORMNUD,DAD>
gif0: flags=8010<POINTOPOINT,MULTICAST> mtu 1280
stf0: flags=0<> mtu 1280
en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500
	ether 11:22:33:44:55:66
	inet6 fe80::abcd:1234:5678:ef01%en0 prefixlen 64 secured scopeid 0x4
	inet 192.168.1.100 netmask 255.255.255.0 broadcast 192.168.1.255
	nd6 options=201<PERFORMNUD,DAD>
	media: autoselect (1000baseT <full-duplex>)
	status: active
utun0: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1380
	inet6 fe80::1234:5678:abcd:ef01%utun0 prefixlen 64 scopeid 0x5
	nd6 options=201<PERFORMNUD,DAD>
utun1: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 2000
	inet6 fe80::abcd:5678:ef01:1234%utun1 prefixlen 64 scopeid 0x6
	nd6 options=201<PERFORMNUD,DAD>`
)

func TestParseNetstatLineHeader(t *testing.T) {
	stat, linkIkd, err := parseNetstatLine(`Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll Drop`)
	assert.Nil(t, linkIkd)
	assert.Nil(t, stat)
	assert.Error(t, err)
	assert.Equal(t, errNetstatHeader, err)
}

func assertLoopbackStat(t *testing.T, err error, stat *IOCountersStat) {
	assert.NoError(t, err)
	assert.Equal(t, uint64(869107), stat.PacketsRecv)
	assert.Equal(t, uint64(0), stat.Errin)
	assert.Equal(t, uint64(169411755), stat.BytesRecv)
	assert.Equal(t, uint64(869108), stat.PacketsSent)
	assert.Equal(t, uint64(1), stat.Errout)
	assert.Equal(t, uint64(169411756), stat.BytesSent)
}

func TestParseNetstatLineLink(t *testing.T) {
	stat, linkID, err := parseNetstatLine(
		`lo0   16384 <Link#1>                        869107     0  169411755   869108     1  169411756     0   0`,
	)
	assertLoopbackStat(t, err, stat)
	assert.NotNil(t, linkID)
	assert.Equal(t, uint(1), *linkID)
}

func TestParseNetstatLineIPv6(t *testing.T) {
	stat, linkID, err := parseNetstatLine(
		`lo0   16384 ::1/128     ::1                 869107     -  169411755   869108     1  169411756     -   -`,
	)
	assertLoopbackStat(t, err, stat)
	assert.Nil(t, linkID)
}

func TestParseNetstatLineIPv4(t *testing.T) {
	stat, linkID, err := parseNetstatLine(
		`lo0   16384 127           127.0.0.1         869107     -  169411755   869108     1  169411756     -   -`,
	)
	assertLoopbackStat(t, err, stat)
	assert.Nil(t, linkID)
}

func TestParseNetstatOutput(t *testing.T) {
	nsInterfaces, err := parseNetstatOutput(netstatNotTruncated)
	assert.NoError(t, err)
	assert.Len(t, nsInterfaces, 8)
	for index := range nsInterfaces {
		assert.NotNil(t, nsInterfaces[index].stat, "Index %d", index)
	}

	assert.NotNil(t, nsInterfaces[0].linkID)
	assert.Equal(t, uint(1), *nsInterfaces[0].linkID)

	assert.Nil(t, nsInterfaces[1].linkID)
	assert.Nil(t, nsInterfaces[2].linkID)
	assert.Nil(t, nsInterfaces[3].linkID)

	assert.NotNil(t, nsInterfaces[4].linkID)
	assert.Equal(t, uint(2), *nsInterfaces[4].linkID)

	assert.NotNil(t, nsInterfaces[5].linkID)
	assert.Equal(t, uint(3), *nsInterfaces[5].linkID)

	assert.NotNil(t, nsInterfaces[6].linkID)
	assert.Equal(t, uint(4), *nsInterfaces[6].linkID)

	assert.Nil(t, nsInterfaces[7].linkID)

	mapUsage := newMapInterfaceNameUsage(nsInterfaces)
	assert.False(t, mapUsage.isTruncated())
	assert.Len(t, mapUsage.notTruncated(), 4)
}

func TestParseNetstatTruncated(t *testing.T) {
	nsInterfaces, err := parseNetstatOutput(netstatTruncated)
	assert.NoError(t, err)
	assert.Len(t, nsInterfaces, 11)
	for index := range nsInterfaces {
		assert.NotNil(t, nsInterfaces[index].stat, "Index %d", index)
	}

	const truncatedIface = "utun8"

	assert.NotNil(t, nsInterfaces[6].linkID)
	assert.Equal(t, uint(88), *nsInterfaces[6].linkID)
	assert.Equal(t, truncatedIface, nsInterfaces[6].stat.Name)

	assert.NotNil(t, nsInterfaces[7].linkID)
	assert.Equal(t, uint(90), *nsInterfaces[7].linkID)
	assert.Equal(t, truncatedIface, nsInterfaces[7].stat.Name)

	assert.NotNil(t, nsInterfaces[8].linkID)
	assert.Equal(t, uint(92), *nsInterfaces[8].linkID)
	assert.Equal(t, truncatedIface, nsInterfaces[8].stat.Name)

	assert.NotNil(t, nsInterfaces[9].linkID)
	assert.Equal(t, uint(93), *nsInterfaces[9].linkID)
	assert.Equal(t, truncatedIface, nsInterfaces[9].stat.Name)

	assert.NotNil(t, nsInterfaces[10].linkID)
	assert.Equal(t, uint(95), *nsInterfaces[10].linkID)
	assert.Equal(t, truncatedIface, nsInterfaces[10].stat.Name)

	mapUsage := newMapInterfaceNameUsage(nsInterfaces)
	assert.True(t, mapUsage.isTruncated())
	assert.Equal(t, 3, len(mapUsage.notTruncated()), "en0, gif0 and stf0")
}

func TestParseIfconfigOutput(t *testing.T) {
	testStats := []IOCountersStat{
		{Name: "lo0"},
		{Name: "en0"},
	}
	err := parseIfconfigOutput(ifconfigOutput, testStats)
	assert.NoError(t, err)
	assert.Len(t, testStats, 2)

	assert.NotNil(t, testStats[0].TransmitSpeed)
	assert.Equal(t, uint64(0), testStats[0].TransmitSpeed)
	assert.NotNil(t, testStats[0].ReceiveSpeed)
	assert.Equal(t, uint64(0), testStats[0].ReceiveSpeed)

	assert.NotNil(t, testStats[1].TransmitSpeed)
	assert.Equal(t, uint64(1000), testStats[1].TransmitSpeed)
	assert.NotNil(t, testStats[1].ReceiveSpeed)
	assert.Equal(t, uint64(1000), testStats[1].ReceiveSpeed)
}
