// Package qry [general utility functions for the dnsblast package]
package qry_test

import (
	"testing"

	"github.com/Arriven/db1000n/src/core/dnsblast/qry"
)

func BenchmarkQtype(b *testing.B) {
	var res uint16

	for n := 0; n < b.N; n++ {
		res |= qry.Qtype("None")
		res |= qry.Qtype("A")
		res |= qry.Qtype("NS")
		res |= qry.Qtype("MD")
		res |= qry.Qtype("MF")
		res |= qry.Qtype("CNAME")
		res |= qry.Qtype("SOA")
		res |= qry.Qtype("MB")
		res |= qry.Qtype("MG")
		res |= qry.Qtype("MR")
		res |= qry.Qtype("NULL")
		res |= qry.Qtype("PTR")
		res |= qry.Qtype("HINFO")
		res |= qry.Qtype("MINFO")
		res |= qry.Qtype("MX")
		res |= qry.Qtype("TXT")
		res |= qry.Qtype("RP")
		res |= qry.Qtype("AFSDB")
		res |= qry.Qtype("X25")
		res |= qry.Qtype("ISDN")
		res |= qry.Qtype("RT")
		res |= qry.Qtype("NSAPPTR")
		res |= qry.Qtype("SIG")
		res |= qry.Qtype("KEY")
		res |= qry.Qtype("PX")
		res |= qry.Qtype("GPOS")
		res |= qry.Qtype("AAAA")
		res |= qry.Qtype("LOC")
		res |= qry.Qtype("NXT")
		res |= qry.Qtype("EID")
		res |= qry.Qtype("NIMLOC")
		res |= qry.Qtype("SRV")
		res |= qry.Qtype("ATMA")
		res |= qry.Qtype("NAPTR")
		res |= qry.Qtype("KX")
		res |= qry.Qtype("CERT")
		res |= qry.Qtype("DNAME")
		res |= qry.Qtype("OPT")
		res |= qry.Qtype("DS")
		res |= qry.Qtype("SSHFP")
		res |= qry.Qtype("RRSIG")
		res |= qry.Qtype("NSEC")
		res |= qry.Qtype("DNSKEY")
		res |= qry.Qtype("DHCID")
		res |= qry.Qtype("NSEC3")
		res |= qry.Qtype("NSEC3PARAM")
		res |= qry.Qtype("TLSA")
		res |= qry.Qtype("SMIMEA")
		res |= qry.Qtype("HIP")
		res |= qry.Qtype("NINFO")
		res |= qry.Qtype("RKEY")
		res |= qry.Qtype("TALINK")
		res |= qry.Qtype("CDS")
		res |= qry.Qtype("CDNSKEY")
		res |= qry.Qtype("OPENPGPKEY")
		res |= qry.Qtype("CSYNC")
		res |= qry.Qtype("SPF")
		res |= qry.Qtype("UINFO")
		res |= qry.Qtype("UID")
		res |= qry.Qtype("GID")
		res |= qry.Qtype("UNSPEC")
		res |= qry.Qtype("NID")
		res |= qry.Qtype("L32")
		res |= qry.Qtype("L64")
		res |= qry.Qtype("LP")
		res |= qry.Qtype("EUI48")
		res |= qry.Qtype("EUI64")
		res |= qry.Qtype("URI")
		res |= qry.Qtype("CAA")
		res |= qry.Qtype("AVC")
		res |= qry.Qtype("default")
	}

	// Avoid optimizing calls away
	if res == 0 {
		b.Fatalf("res is 0")
	}
}
