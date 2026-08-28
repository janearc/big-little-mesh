// Package leaseproto exposes the lease.v1 .proto source text for Confluent
// Schema-Registry registration, following the observability.v1 idiom: the
// registered wire schema and the generated bindings come from one place.
package leaseproto

import _ "embed"

// Schema is the raw lease.proto, embedded for Schema-Registry registration
// (RecordNameStrategy subject lease.v1.LeaseVerdict).
//
//go:embed lease.proto
var Schema string

// SubjectLeaseVerdict is the canonical RecordNameStrategy subject, co-located
// with the schema it registers under for the same reason observability.v1 does
// it: one home, no scattered literals to desync.
const SubjectLeaseVerdict = "lease.v1.LeaseVerdict"
