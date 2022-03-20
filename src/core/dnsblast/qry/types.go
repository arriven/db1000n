// Package qry [general utility functions for the dnsblast package]
package qry

import "github.com/miekg/dns"

// ResponseCode convert numerical response code value to string
func ResponseCode(rc int) string {
	rcodes := map[int]string{
		0: "NOERROR",
		1: "FORMERR",
		2: "SERVFAIL",
		3: "NXDOMAIN",
		4: "NOTIMP",
		5: "REFUSED",
		6: "YXDOMAIN",
		7: "YXRRSET",
		8: "NOTAUTH",
		9: "NOTZONE",
	}

	return rcodes[rc]
}

// Qtype is used to conver string representation of query type into proper dns format
func Qtype(qt string) uint16 {
	res, ok := map[string]uint16{
		"None":       dns.TypeNone,
		"A":          dns.TypeA,
		"NS":         dns.TypeNS,
		"MD":         dns.TypeMD,
		"MF":         dns.TypeMF,
		"CNAME":      dns.TypeCNAME,
		"SOA":        dns.TypeSOA,
		"MB":         dns.TypeMB,
		"MG":         dns.TypeMG,
		"MR":         dns.TypeMR,
		"NULL":       dns.TypeNULL,
		"PTR":        dns.TypePTR,
		"HINFO":      dns.TypeHINFO,
		"MINFO":      dns.TypeMINFO,
		"MX":         dns.TypeMX,
		"TXT":        dns.TypeTXT,
		"RP":         dns.TypeRP,
		"AFSDB":      dns.TypeAFSDB,
		"X25":        dns.TypeX25,
		"ISDN":       dns.TypeISDN,
		"RT":         dns.TypeRT,
		"NSAPPTR":    dns.TypeNSAPPTR,
		"SIG":        dns.TypeSIG,
		"KEY":        dns.TypeKEY,
		"PX":         dns.TypePX,
		"GPOS":       dns.TypeGPOS,
		"AAAA":       dns.TypeAAAA,
		"LOC":        dns.TypeLOC,
		"NXT":        dns.TypeNXT,
		"EID":        dns.TypeEID,
		"NIMLOC":     dns.TypeNIMLOC,
		"SRV":        dns.TypeSRV,
		"ATMA":       dns.TypeATMA,
		"NAPTR":      dns.TypeNAPTR,
		"KX":         dns.TypeKX,
		"CERT":       dns.TypeCERT,
		"DNAME":      dns.TypeDNAME,
		"OPT":        dns.TypeOPT,
		"DS":         dns.TypeDS,
		"SSHFP":      dns.TypeSSHFP,
		"RRSIG":      dns.TypeRRSIG,
		"NSEC":       dns.TypeNSEC,
		"DNSKEY":     dns.TypeDNSKEY,
		"DHCID":      dns.TypeDHCID,
		"NSEC3":      dns.TypeNSEC3,
		"NSEC3PARAM": dns.TypeNSEC3PARAM,
		"TLSA":       dns.TypeTLSA,
		"SMIMEA":     dns.TypeSMIMEA,
		"HIP":        dns.TypeHIP,
		"NINFO":      dns.TypeNINFO,
		"RKEY":       dns.TypeRKEY,
		"TALINK":     dns.TypeTALINK,
		"CDS":        dns.TypeCDS,
		"CDNSKEY":    dns.TypeCDNSKEY,
		"OPENPGPKEY": dns.TypeOPENPGPKEY,
		"CSYNC":      dns.TypeCSYNC,
		"SPF":        dns.TypeSPF,
		"UINFO":      dns.TypeUINFO,
		"UID":        dns.TypeUID,
		"GID":        dns.TypeGID,
		"UNSPEC":     dns.TypeUNSPEC,
		"NID":        dns.TypeNID,
		"L32":        dns.TypeL32,
		"L64":        dns.TypeL64,
		"LP":         dns.TypeLP,
		"EUI48":      dns.TypeEUI48,
		"EUI64":      dns.TypeEUI64,
		"URI":        dns.TypeURI,
		"CAA":        dns.TypeCAA,
		"AVC":        dns.TypeAVC,
	}[qt]
	if !ok {
		return dns.TypeA
	}

	return res
}
