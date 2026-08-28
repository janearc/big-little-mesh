import datetime

from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from lease.v1 import lease_pb2 as _lease_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FindingClass(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FINDING_CLASS_UNSPECIFIED: _ClassVar[FindingClass]
    FINDING_CLASS_REFUSAL: _ClassVar[FindingClass]

class FindingKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FINDING_KIND_UNSPECIFIED: _ClassVar[FindingKind]
    FINDING_KIND_VOID: _ClassVar[FindingKind]
    FINDING_KIND_SILENT: _ClassVar[FindingKind]
    FINDING_KIND_OFF_CONTRACT: _ClassVar[FindingKind]
    FINDING_KIND_LEASE_EXPIRED: _ClassVar[FindingKind]
FINDING_CLASS_UNSPECIFIED: FindingClass
FINDING_CLASS_REFUSAL: FindingClass
FINDING_KIND_UNSPECIFIED: FindingKind
FINDING_KIND_VOID: FindingKind
FINDING_KIND_SILENT: FindingKind
FINDING_KIND_OFF_CONTRACT: FindingKind
FINDING_KIND_LEASE_EXPIRED: FindingKind

class TopicRow(_message.Message):
    __slots__ = ("topic", "last_record", "consumers", "void", "silent", "expected_cadence", "silent_for", "off_contract_records", "last_off_contract")
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    LAST_RECORD_FIELD_NUMBER: _ClassVar[int]
    CONSUMERS_FIELD_NUMBER: _ClassVar[int]
    VOID_FIELD_NUMBER: _ClassVar[int]
    SILENT_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_CADENCE_FIELD_NUMBER: _ClassVar[int]
    SILENT_FOR_FIELD_NUMBER: _ClassVar[int]
    OFF_CONTRACT_RECORDS_FIELD_NUMBER: _ClassVar[int]
    LAST_OFF_CONTRACT_FIELD_NUMBER: _ClassVar[int]
    topic: str
    last_record: _timestamp_pb2.Timestamp
    consumers: _containers.RepeatedScalarFieldContainer[str]
    void: bool
    silent: bool
    expected_cadence: _duration_pb2.Duration
    silent_for: _duration_pb2.Duration
    off_contract_records: int
    last_off_contract: _timestamp_pb2.Timestamp
    def __init__(self, topic: _Optional[str] = ..., last_record: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., consumers: _Optional[_Iterable[str]] = ..., void: _Optional[bool] = ..., silent: _Optional[bool] = ..., expected_cadence: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., silent_for: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., off_contract_records: _Optional[int] = ..., last_off_contract: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Finding(_message.Message):
    __slots__ = ("kind", "topic", "detail")
    CLASS_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    DETAIL_FIELD_NUMBER: _ClassVar[int]
    kind: FindingKind
    topic: str
    detail: str
    def __init__(self, kind: _Optional[_Union[FindingKind, str]] = ..., topic: _Optional[str] = ..., detail: _Optional[str] = ..., **kwargs) -> None: ...

class TopicList(_message.Message):
    __slots__ = ("topics",)
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    topics: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, topics: _Optional[_Iterable[str]] = ...) -> None: ...

class Report(_message.Message):
    __slots__ = ("service", "generated_at", "authorized", "topics", "groups", "findings")
    class GroupsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: TopicList
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[TopicList, _Mapping]] = ...) -> None: ...
    SERVICE_FIELD_NUMBER: _ClassVar[int]
    GENERATED_AT_FIELD_NUMBER: _ClassVar[int]
    AUTHORIZED_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    GROUPS_FIELD_NUMBER: _ClassVar[int]
    FINDINGS_FIELD_NUMBER: _ClassVar[int]
    service: str
    generated_at: _timestamp_pb2.Timestamp
    authorized: _containers.RepeatedCompositeFieldContainer[_lease_pb2.LeaseVerdict]
    topics: _containers.RepeatedCompositeFieldContainer[TopicRow]
    groups: _containers.MessageMap[str, TopicList]
    findings: _containers.RepeatedCompositeFieldContainer[Finding]
    def __init__(self, service: _Optional[str] = ..., generated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., authorized: _Optional[_Iterable[_Union[_lease_pb2.LeaseVerdict, _Mapping]]] = ..., topics: _Optional[_Iterable[_Union[TopicRow, _Mapping]]] = ..., groups: _Optional[_Mapping[str, TopicList]] = ..., findings: _Optional[_Iterable[_Union[Finding, _Mapping]]] = ...) -> None: ...
