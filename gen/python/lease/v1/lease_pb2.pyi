import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LeaseState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LEASE_STATE_UNSPECIFIED: _ClassVar[LeaseState]
    LEASE_STATE_AUTHORIZED: _ClassVar[LeaseState]
    LEASE_STATE_EXPIRING: _ClassVar[LeaseState]
    LEASE_STATE_EXPIRED: _ClassVar[LeaseState]
LEASE_STATE_UNSPECIFIED: LeaseState
LEASE_STATE_AUTHORIZED: LeaseState
LEASE_STATE_EXPIRING: LeaseState
LEASE_STATE_EXPIRED: LeaseState

class LeaseVerdict(_message.Message):
    __slots__ = ("service_name", "state", "last_heartbeat", "expires_at", "issued_at", "authority")
    SERVICE_NAME_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    LAST_HEARTBEAT_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    ISSUED_AT_FIELD_NUMBER: _ClassVar[int]
    AUTHORITY_FIELD_NUMBER: _ClassVar[int]
    service_name: str
    state: LeaseState
    last_heartbeat: _timestamp_pb2.Timestamp
    expires_at: _timestamp_pb2.Timestamp
    issued_at: _timestamp_pb2.Timestamp
    authority: str
    def __init__(self, service_name: _Optional[str] = ..., state: _Optional[_Union[LeaseState, str]] = ..., last_heartbeat: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., expires_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., issued_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., authority: _Optional[str] = ...) -> None: ...
